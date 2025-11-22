// CRC: crc-P2PWebAppClient.md, Spec: main.md
import {
  Message,
  StringResponse,
  PeerResponse,
  ListPeersResponse,
  FileEntry,
  FileContent,
  StoreFileResponse,
  ProtocolDataCallback,
  TopicDataCallback,
  PeerChangeCallback,
  PeerDataRequest,
  TopicDataRequest,
  PeerChangeRequest,
  AckRequest,
  PeerFilesRequest,
  GotFileRequest,
} from './types.js';

interface PendingRequest {
  resolve: (value: any) => void;
  reject: (error: Error) => void;
}

interface PendingPromiseRequest<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (error: Error) => void;
}

// CRC: crc-P2PWebAppClient.md
export class P2PWebAppClient {
  private ws: WebSocket | null = null;
  private _peerID: string | null = null;
  private _peerKey: string | null = null;
  private requestID: number = 0;
  private pending: Map<number, PendingRequest> = new Map();
  private protocolListeners: Map<string, ProtocolDataCallback> = new Map(); // key: protocol
  private topicListeners: Map<string, TopicDataCallback> = new Map();
  private peerChangeListeners: Map<string, PeerChangeCallback> = new Map(); // key: topic

  // Message queuing for sequential processing
  private messageQueue: Message[] = [];
  private processingMessage: boolean = false;

  // Ack tracking for send() promises
  private nextAckNumber: number = 0;
  private ackPending: Map<number, PendingRequest> = new Map(); // key: ack number

  // File operation promise tracking
  private fileListPending: Map<string, PendingPromiseRequest<{ rootCID: string; entries: { [path: string]: FileEntry } }>> = new Map(); // key: peerID
  private getFilePending: Map<string, PendingPromiseRequest<FileContent>> = new Map(); // key: CID

  /**
   * Connect to the WebSocket server and initialize peer identity
   * @param peerKey Optional peer key to restore previous identity
   * @returns Promise resolving to [peerID, peerKey] tuple
   * CRC: crc-P2PWebAppClient.md
   */
  async connect(peerKey?: string): Promise<[string, string]> {
    const wsUrl = this.getDefaultWSUrl();

    // First, establish WebSocket connection
    await new Promise<void>((resolve, reject) => {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => resolve();
      this.ws.onerror = (error) => reject(new Error('WebSocket connection failed'));
      this.ws.onmessage = (event) => this.handleMessage(event);
      this.ws.onclose = () => this.handleClose();
    });

    // Then, initialize peer identity
    const result = await this.sendRequest('peer', peerKey ? { peerkey: peerKey } : {});
    const response = result as PeerResponse;
    this._peerID = response.peerid;
    this._peerKey = response.peerkey;
    return [this._peerID, this._peerKey];
  }

  /**
   * Close the WebSocket connection
   */
  close(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Start a protocol with a data listener (required before sending)
   * The listener receives (peer, data) for all messages on this protocol
   */
  async start(protocol: string, onData: ProtocolDataCallback): Promise<void> {
    if (this.protocolListeners.has(protocol)) {
      throw new Error(`Protocol '${protocol}' already started`);
    }
    await this.sendRequest('start', { protocol });
    this.protocolListeners.set(protocol, onData);
  }

  /**
   * Stop a protocol and remove its listener
   */
  async stop(protocol: string): Promise<void> {
    if (!this.protocolListeners.has(protocol)) {
      throw new Error(`Protocol '${protocol}' not started`);
    }
    await this.sendRequest('stop', { protocol });
    this.protocolListeners.delete(protocol);
  }

  /**
   * Send data to a peer on a protocol
   * @param peer Target peer ID
   * @param protocol Protocol name
   * @param data Data to send
   * @returns Promise that resolves when delivery is confirmed
   */
  async send(peer: string, protocol: string, data: any): Promise<void> {
    if (!this.protocolListeners.has(protocol)) {
      throw new Error(`Cannot send on protocol '${protocol}': protocol not started. Call start() first.`);
    }

    // Always request acknowledgment
    const ackNum = this.nextAckNumber++;

    // Create promise that will resolve when ack is received
    const ackPromise = new Promise<void>((resolve, reject) => {
      this.ackPending.set(ackNum, { resolve, reject });
    });

    // Send the request
    await this.sendRequest('send', { peer, protocol, data, ack: ackNum });

    // Wait for ack
    return ackPromise;
  }

  /**
   * Subscribe to a topic with data listener and optional peer change listener
   * Automatically monitors the topic for peer join/leave events if onPeerChange is provided
   */
  async subscribe(topic: string, onData: TopicDataCallback, onPeerChange?: PeerChangeCallback): Promise<void> {
    this.topicListeners.set(topic, onData);
    if (onPeerChange) {
      this.peerChangeListeners.set(topic, onPeerChange);
    }
    await this.sendRequest('subscribe', { topic });
  }

  /**
   * Publish data to a topic
   */
  async publish(topic: string, data: any): Promise<void> {
    await this.sendRequest('publish', { topic, data });
  }

  /**
   * Unsubscribe from a topic and stop monitoring peer changes
   */
  async unsubscribe(topic: string): Promise<void> {
    this.topicListeners.delete(topic);
    this.peerChangeListeners.delete(topic);
    await this.sendRequest('unsubscribe', { topic });
  }

  /**
   * List peers subscribed to a topic
   */
  async listPeers(topic: string): Promise<string[]> {
    const result = await this.sendRequest('listpeers', { topic });
    const response = result as ListPeersResponse;
    return response.peers || [];
  }

  /**
   * List files for a peer
   * @param peerid Peer ID whose files to list
   * @returns Promise resolving with {rootCID, entries}
   */
  async listFiles(peerid: string): Promise<{ rootCID: string; entries: { [path: string]: FileEntry } }> {
    // Check if there's already a pending request for this peer
    if (this.fileListPending.has(peerid)) {
      // Wait for existing request to complete
      return this.fileListPending.get(peerid)!.promise;
    }

    // Create promise that will resolve when peerFiles message is received
    let resolveFunc: (value: { rootCID: string; entries: { [path: string]: FileEntry } }) => void;
    let rejectFunc: (error: Error) => void;

    const promise = new Promise<{ rootCID: string; entries: { [path: string]: FileEntry } }>((resolve, reject) => {
      resolveFunc = resolve;
      rejectFunc = reject;
    });

    this.fileListPending.set(peerid, { promise, resolve: resolveFunc!, reject: rejectFunc! });

    // Send request (actual result comes via peerFiles server message)
    await this.sendRequest('listfiles', { peerid });

    return promise;
  }

  /**
   * Get file or directory content by CID
   * @param cid Content identifier
   * @returns Promise resolving with file content or rejecting on error
   */
  async getFile(cid: string, fallbackPeerID?: string): Promise<FileContent> {
    // Check if there's already a pending request for this CID
    if (this.getFilePending.has(cid)) {
      // Wait for existing request to complete
      return this.getFilePending.get(cid)!.promise;
    }

    // Create promise that will resolve when gotFile message is received
    let resolveFunc: (value: FileContent) => void;
    let rejectFunc: (error: Error) => void;

    const promise = new Promise<FileContent>((resolve, reject) => {
      resolveFunc = resolve;
      rejectFunc = reject;
    });

    this.getFilePending.set(cid, { promise, resolve: resolveFunc!, reject: rejectFunc! });

    // Send request (actual result comes via gotFile server message)
    const params: any = { cid };
    if (fallbackPeerID) {
      params.fallbackPeerID = fallbackPeerID;
    }
    await this.sendRequest('getfile', params);

    return promise;
  }

  /**
   * Store file for this peer
   * @param path File path identifier
   * @param content File content as string or Uint8Array
   * @returns Promise resolving to StoreFileResponse with fileCid and rootCid
   */
  async storeFile(path: string, content: string | Uint8Array): Promise<StoreFileResponse> {
    let base64Content: string;

    if (typeof content === 'string') {
      // Convert string to UTF-8 bytes then to base64
      const encoder = new TextEncoder();
      const bytes = encoder.encode(content);
      let binaryString = '';
      for (let i = 0; i < bytes.length; i++) {
        binaryString += String.fromCharCode(bytes[i]);
      }
      base64Content = btoa(binaryString);
    } else {
      // Encode Uint8Array to base64 for binary files
      let binaryString = '';
      for (let i = 0; i < content.length; i++) {
        binaryString += String.fromCharCode(content[i]);
      }
      base64Content = btoa(binaryString);
    }

    const result = await this.sendRequest('storefile', { path, content: base64Content, directory: false });
    return { fileCid: result.fileCid, rootCid: result.rootCid };
  }

  /**
   * Create directory for this peer
   * @param path Directory path identifier
   * @returns Promise resolving to StoreFileResponse with fileCid and rootCid
   */
  async createDirectory(path: string): Promise<StoreFileResponse> {
    const result = await this.sendRequest('storefile', { path, content: undefined, directory: true });
    return { fileCid: result.fileCid, rootCid: result.rootCid };
  }

  /**
   * Remove a file from this peer's storage
   * @param path File path identifier
   */
  async removeFile(path: string): Promise<void> {
    await this.sendRequest('removefile', { path });
  }

  /**
   * Get the current peer ID
   */
  get peerID(): string | null {
    return this._peerID;
  }

  /**
   * Get the current peer key
   */
  get peerKey(): string | null {
    return this._peerKey;
  }

  // Private methods

  private getDefaultWSUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}/ws`;
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const msg: Message = JSON.parse(event.data);

      if (msg.isresponse) {
        // Handle response to our request (not queued, processed immediately)
        const pending = this.pending.get(msg.requestid);
        if (pending) {
          this.pending.delete(msg.requestid);
          if (msg.error) {
            pending.reject(new Error(msg.error.message));
          } else {
            pending.resolve(msg.result);
          }
        }
      } else {
        // Queue server-initiated requests for sequential processing
        this.messageQueue.push(msg);
        this.processMessageQueue();
      }
    } catch (error) {
      console.error('Failed to handle message:', error);
    }
  }

  private async processMessageQueue(): Promise<void> {
    // If already processing, return (next message will be processed when current finishes)
    if (this.processingMessage) {
      return;
    }

    this.processingMessage = true;

    try {
      while (this.messageQueue.length > 0) {
        const msg = this.messageQueue.shift();
        if (msg) {
          try {
            await this.handleServerRequest(msg);
          } catch (error) {
            console.error('Error processing server message:', error);
            // Continue processing next message despite error
          }
        }
      }
    } finally {
      this.processingMessage = false;
    }
  }

  private async handleServerRequest(msg: Message): Promise<void> {
    switch (msg.method) {
      case 'peerData':
        if (msg.params) {
          const req = msg.params as PeerDataRequest;
          const listener = this.protocolListeners.get(req.protocol);
          if (listener) {
            try {
              await listener(req.peer, req.data);
            } catch (error) {
              console.error('Error in peerData listener:', error);
            }
          }
        }
        break;

      case 'topicData':
        if (msg.params) {
          const req = msg.params as TopicDataRequest;
          const listener = this.topicListeners.get(req.topic);
          if (listener) {
            try {
              await listener(req.peerid, req.data);
            } catch (error) {
              console.error('Error in topicData listener:', error);
            }
          }
        }
        break;

      case 'peerChange':
        if (msg.params) {
          const req = msg.params as PeerChangeRequest;
          const listener = this.peerChangeListeners.get(req.topic);
          if (listener) {
            try {
              await listener(req.peerid, req.joined);
            } catch (error) {
              console.error('Error in peerChange listener:', error);
            }
          }
        }
        break;

      case 'ack':
        if (msg.params) {
          const req = msg.params as AckRequest;
          const pending = this.ackPending.get(req.ack);
          if (pending) {
            this.ackPending.delete(req.ack); // Remove pending promise after use
            pending.resolve(undefined);
          }
        }
        break;

      case 'peerFiles':
        if (msg.params) {
          const req = msg.params as PeerFilesRequest;
          const pending = this.fileListPending.get(req.peerid);
          if (pending) {
            this.fileListPending.delete(req.peerid); // Remove pending promise after use
            pending.resolve({ rootCID: req.cid, entries: req.entries });
          }
        }
        break;

      case 'gotFile':
        if (msg.params) {
          const req = msg.params as GotFileRequest;
          const pending = this.getFilePending.get(req.cid);
          if (pending) {
            this.getFilePending.delete(req.cid); // Remove pending promise after use
            if (req.success) {
              pending.resolve(req.content as FileContent);
            } else {
              pending.reject(new Error(req.content?.error || 'Failed to retrieve file'));
            }
          }
        }
        break;
    }
  }

  private handleClose(): void {
    // Clean up all listeners on disconnect
    this.protocolListeners.clear();
    this.topicListeners.clear();
    this.peerChangeListeners.clear();

    // Reject all pending promises
    this.ackPending.forEach(pending => pending.reject(new Error('Connection closed')));
    this.ackPending.clear();

    this.fileListPending.forEach(pending => pending.reject(new Error('Connection closed')));
    this.fileListPending.clear();

    this.getFilePending.forEach(pending => pending.reject(new Error('Connection closed')));
    this.getFilePending.clear();

    this.messageQueue.length = 0;
    this.processingMessage = false;
  }

  private sendRequest(method: string, params: any): Promise<any> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return Promise.reject(new Error('WebSocket not connected'));
    }

    return new Promise((resolve, reject) => {
      const id = this.requestID++;
      this.pending.set(id, { resolve, reject });

      const msg: Message = {
        requestid: id,
        method,
        params,
        isresponse: false,
      };

      this.ws!.send(JSON.stringify(msg));
    });
  }
}

/**
 * Convenience function to create and connect a P2PWebAppClient in one call
 * @param peerKey Optional peer key to restore previous identity
 * @returns Promise resolving to connected P2PWebAppClient instance
 */
export async function connect(peerKey?: string): Promise<P2PWebAppClient> {
  const client = new P2PWebAppClient();
  await client.connect(peerKey);
  return client;
}
