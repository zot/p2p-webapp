// Message types matching the Go protocol definitions

export interface Message {
  requestid: number;
  method?: string;
  params?: any;
  result?: any;
  error?: ErrorResponse;
  isresponse: boolean;
}

export interface ErrorResponse {
  code: number;
  message: string;
}

export interface StringResponse {
  value: string;
}

export interface PeerResponse {
  peerid: string;
  peerkey: string;
  version: string;
}

// Client request message types

export interface ConnectOptions {
  peerKey?: string;
  onClose?: () => void;
}

export interface PeerRequest {
  peerkey?: string;
  rootDirectory?: string; // Optional CID of peer's root directory
}

export interface StartRequest {
  protocol: string;
}

export interface StopRequest {
  protocol: string;
}

export interface SendRequest {
  peer: string;
  protocol: string;
  data: any;
  ack: number; // -1 = no ack, >= 0 = request ack with this number
}

export interface SubscribeRequest {
  topic: string;
}

export interface PublishRequest {
  topic: string;
  data: any;
}

export interface UnsubscribeRequest {
  topic: string;
}

export interface ListPeersRequest {
  topic: string;
}

export interface ListPeersResponse {
  peers: string[];
}

// File operation types

export interface FileEntry {
  type: "file" | "directory";
  cid: string;
  mimeType?: string; // Only for files
}

export interface ListFilesRequest {
  peerid: string; // Peer whose files to list
}

export interface GetFileRequest {
  cid: string;
}

export interface StoreFileRequest {
  path: string;
  content?: string; // base64 encoded file content (null for directories)
  directory: boolean; // true = create directory, false = create file
}

export interface RemoveFileRequest {
  path: string;
}

export interface StoreFileResponse {
  fileCid: string; // CID of the stored file/directory node
  rootCid: string; // CID of the peer's updated root directory
}

// Server request message types

export interface PeerDataRequest {
  peer: string;
  protocol: string;
  data: any;
}

export interface TopicDataRequest {
  topic: string;
  peerid: string;
  data: any;
}

export interface PeerChangeRequest {
  topic: string;
  peerid: string;
  joined: boolean; // true = joined, false = left
}

export interface AckRequest {
  ack: number;
}

export interface PeerFilesRequest {
  peerid: string; // Target peer whose files were listed
  cid: string; // Root directory CID
  entries: { [path: string]: FileEntry }; // Full pathname tree
}

export interface GotFileRequest {
  cid: string; // Requested CID
  success: boolean; // Whether retrieval was successful
  content: any; // File content or error info
}

// Callback types

// Callbacks can be sync or async for flexibility
export type ProtocolDataCallback = (peer: string, data: any) => void | Promise<void>;
export type TopicDataCallback = (peerID: string, data: any) => void | Promise<void>;
export type PeerChangeCallback = (peerID: string, joined: boolean) => void | Promise<void>;

// File content types
export type FileContent = FileContentFile | FileContentDirectory;

export interface FileContentFile {
  type: 'file';
  mimeType: string;
  content: string; // base64-encoded
}

export interface FileContentDirectory {
  type: 'directory';
  entries: {
    [pathname: string]: string; // CID
  };
}
