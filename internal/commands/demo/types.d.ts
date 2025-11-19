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
}
export interface PeerRequest {
    peerkey?: string;
    rootDirectory?: string;
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
    ack: number;
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
export interface FileEntry {
    type: "file" | "directory";
    cid: string;
    mimeType?: string;
}
export interface ListFilesRequest {
    peerid: string;
}
export interface GetFileRequest {
    cid: string;
}
export interface StoreFileRequest {
    path: string;
    content?: string;
    directory: boolean;
}
export interface RemoveFileRequest {
    path: string;
}
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
    joined: boolean;
}
export interface AckRequest {
    ack: number;
}
export interface PeerFilesRequest {
    peerid: string;
    cid: string;
    entries: {
        [path: string]: FileEntry;
    };
}
export interface GotFileRequest {
    cid: string;
    success: boolean;
    content: any;
}
export type ProtocolDataCallback = (peer: string, data: any) => void | Promise<void>;
export type TopicDataCallback = (peerID: string, data: any) => void | Promise<void>;
export type PeerChangeCallback = (peerID: string, joined: boolean) => void | Promise<void>;
export type FileContent = FileContentFile | FileContentDirectory;
export interface FileContentFile {
    type: 'file';
    mimeType: string;
    content: string;
}
export interface FileContentDirectory {
    type: 'directory';
    entries: {
        [pathname: string]: string;
    };
}
