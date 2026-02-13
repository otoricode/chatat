// Type declarations for native backup modules
// These modules require native setup — see INTEGRATION.md

declare module '@react-native-google-signin/google-signin' {
  export const GoogleSignin: {
    configure(options: { scopes?: string[] }): void;
    hasPlayServices(): Promise<void>;
    signIn(): Promise<{ user: { email: string; name: string } }>;
    getTokens(): Promise<{ accessToken: string; idToken: string }>;
  };
}

declare module '@robinbobin/react-native-google-drive-api-wrapper' {
  interface FileMetadata {
    id: string;
    name: string;
    size?: string;
    createdTime?: string;
    mimeType?: string;
  }

  interface ListResult {
    files: FileMetadata[];
  }

  interface CreateResult {
    result: FileMetadata;
  }

  interface Uploader {
    setData(data: string, mimeType: string): Uploader;
    setRequestBody(body: Record<string, unknown>): Uploader;
    execute(): Promise<FileMetadata>;
  }

  interface MetadataUploader {
    setRequestBody(body: Record<string, unknown>): MetadataUploader;
  }

  interface Files {
    list(params: Record<string, unknown>): Promise<ListResult>;
    getContent(fileId: string): Promise<string>;
    newMultipartUploader(): Uploader;
    newMetadataOnlyUploader(): MetadataUploader;
    createIfNotExists(
      query: Record<string, string>,
      uploader: MetadataUploader,
    ): Promise<CreateResult>;
  }

  export class GDrive {
    accessToken: string;
    files: Files;
  }

  export const MimeTypes: {
    JSON: string;
    FOLDER: string;
  };
}

declare module 'react-native-cloud-store' {
  const CloudStore: {
    isICloudAvailable(): Promise<boolean>;
    writeFile(
      path: string,
      content: string,
      options?: { override?: boolean },
    ): Promise<void>;
    readFile(path: string): Promise<string>;
    readDir(path: string): Promise<string[]>;
    exist(path: string): Promise<boolean>;
  };
  export default CloudStore;
}
