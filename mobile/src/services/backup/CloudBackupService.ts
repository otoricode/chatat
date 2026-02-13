// Cloud Backup Service — Platform-agnostic abstraction for Google Drive and iCloud backup
// Google Drive: Android, iCloud: iOS
// NOTE: Requires native module setup. See INTEGRATION.md for configuration details.

import { Platform } from 'react-native';
import { backupApi } from '@/services/api/backup';
import type { BackupBundle, BackupPlatform } from '@/services/api/backup';

export type BackupFile = {
  id: string;
  name: string;
  sizeBytes: number;
  createdAt: string;
};

export type BackupProgress = {
  phase: 'exporting' | 'uploading' | 'downloading' | 'importing';
  progress: number; // 0 to 1
};

type ProgressCallback = (progress: BackupProgress) => void;

// --- Google Drive Implementation ---

async function googleDriveBackup(onProgress?: ProgressCallback): Promise<void> {
  try {
    // Dynamic import to avoid crash on iOS
    const { GoogleSignin } = await import(
      '@react-native-google-signin/google-signin'
    );
    const { GDrive, MimeTypes } = await import(
      '@robinbobin/react-native-google-drive-api-wrapper'
    );

    // 1. Sign in
    await GoogleSignin.hasPlayServices();
    await GoogleSignin.signIn();
    const tokens = await GoogleSignin.getTokens();

    const gdrive = new GDrive();
    gdrive.accessToken = tokens.accessToken;

    // 2. Export data from server
    onProgress?.({ phase: 'exporting', progress: 0.1 });
    const res = await backupApi.export();
    const apiData = res.data as unknown as { success: boolean; data: BackupBundle };
    const bundle = apiData.data;

    onProgress?.({ phase: 'exporting', progress: 0.5 });

    // 3. Get or create backup folder
    const folderName = 'Chatat Backup';
    let folderId: string | null = null;

    const folderSearch = await gdrive.files.list({
      q: `name='${folderName}' and mimeType='application/vnd.google-apps.folder' and trashed=false`,
      fields: 'files(id)',
    });

    if (folderSearch.files && folderSearch.files.length > 0) {
      const first = folderSearch.files[0];
      if (first) folderId = first.id;
    } else {
      const folder = await gdrive.files.createIfNotExists(
        { q: `name='${folderName}' and mimeType='application/vnd.google-apps.folder'` },
        gdrive.files.newMetadataOnlyUploader().setRequestBody({
          name: folderName,
          mimeType: 'application/vnd.google-apps.folder',
        }),
      );
      folderId = folder.result.id;
    }

    // 4. Upload
    onProgress?.({ phase: 'uploading', progress: 0.6 });
    const backupJson = JSON.stringify(bundle);
    const fileName = `chatat-${new Date().toISOString().split('T')[0]}.json`;

    await gdrive.files
      .newMultipartUploader()
      .setData(backupJson, MimeTypes.JSON)
      .setRequestBody({ name: fileName, parents: [folderId] })
      .execute();

    onProgress?.({ phase: 'uploading', progress: 0.9 });

    // 5. Log backup record
    await backupApi.log({
      sizeBytes: backupJson.length,
      platform: 'google_drive',
      status: 'completed',
    });

    onProgress?.({ phase: 'uploading', progress: 1.0 });
  } catch (error) {
    // Log failed backup
    try {
      await backupApi.log({
        sizeBytes: 0,
        platform: 'google_drive',
        status: 'failed',
      });
    } catch {
      // ignore logging error
    }
    throw error;
  }
}

async function googleDriveRestore(
  onProgress?: ProgressCallback,
): Promise<BackupBundle | null> {
  const { GoogleSignin } = await import(
    '@react-native-google-signin/google-signin'
  );
  const { GDrive } = await import(
    '@robinbobin/react-native-google-drive-api-wrapper'
  );

  await GoogleSignin.hasPlayServices();
  await GoogleSignin.signIn();
  const tokens = await GoogleSignin.getTokens();

  const gdrive = new GDrive();
  gdrive.accessToken = tokens.accessToken;

  // Find backup folder
  const folderName = 'Chatat Backup';
  const folderSearch = await gdrive.files.list({
    q: `name='${folderName}' and mimeType='application/vnd.google-apps.folder' and trashed=false`,
    fields: 'files(id)',
  });

  if (!folderSearch.files?.length) return null;
  const folderFile = folderSearch.files[0];
  if (!folderFile) return null;
  const folderId = folderFile.id;

  // List backup files
  onProgress?.({ phase: 'downloading', progress: 0.2 });
  const filesResult = await gdrive.files.list({
    q: `'${folderId}' in parents and trashed=false`,
    orderBy: 'createdTime desc',
    pageSize: 1,
    fields: 'files(id, name)',
  });

  if (!filesResult.files?.length) return null;

  const latestFile = filesResult.files[0];
  if (!latestFile) return null;

  // Download latest
  onProgress?.({ phase: 'downloading', progress: 0.5 });
  const content = await gdrive.files.getContent(latestFile.id);
  const bundle = JSON.parse(content) as BackupBundle;

  // Import to server
  onProgress?.({ phase: 'importing', progress: 0.7 });
  await backupApi.import(bundle);

  onProgress?.({ phase: 'importing', progress: 1.0 });
  return bundle;
}

async function googleDriveListBackups(): Promise<BackupFile[]> {
  try {
    const { GoogleSignin } = await import(
      '@react-native-google-signin/google-signin'
    );
    const { GDrive } = await import(
      '@robinbobin/react-native-google-drive-api-wrapper'
    );

    await GoogleSignin.hasPlayServices();
    const tokens = await GoogleSignin.getTokens();

    const gdrive = new GDrive();
    gdrive.accessToken = tokens.accessToken;

    const folderName = 'Chatat Backup';
    const folderSearch = await gdrive.files.list({
      q: `name='${folderName}' and mimeType='application/vnd.google-apps.folder' and trashed=false`,
      fields: 'files(id)',
    });

    if (!folderSearch.files?.length) return [];
    const firstFolder = folderSearch.files[0];
    if (!firstFolder) return [];
    const folderId = firstFolder.id;

    const filesResult = await gdrive.files.list({
      q: `'${folderId}' in parents and trashed=false`,
      orderBy: 'createdTime desc',
      pageSize: 10,
      fields: 'files(id, name, size, createdTime)',
    });

    return (
      filesResult.files?.map(
        (f: { id: string; name: string; size?: string; createdTime?: string }) => ({
          id: f.id,
          name: f.name,
          sizeBytes: parseInt(f.size || '0', 10),
          createdAt: f.createdTime || '',
        }),
      ) || []
    );
  } catch {
    return [];
  }
}

// --- iCloud Implementation ---

async function icloudBackup(onProgress?: ProgressCallback): Promise<void> {
  try {
    const CloudStore = (await import('react-native-cloud-store')).default;
    const BACKUP_DIR = 'Documents/Chatat';

    const isAvailable = await CloudStore.isICloudAvailable();
    if (!isAvailable) {
      throw new Error('iCloud is not available');
    }

    // 1. Export data from server
    onProgress?.({ phase: 'exporting', progress: 0.1 });
    const res = await backupApi.export();
    const apiData = res.data as unknown as { success: boolean; data: BackupBundle };
    const bundle = apiData.data;

    onProgress?.({ phase: 'exporting', progress: 0.5 });

    // 2. Write to iCloud
    onProgress?.({ phase: 'uploading', progress: 0.6 });
    const backupJson = JSON.stringify(bundle);
    const fileName = `chatat-${new Date().toISOString().split('T')[0]}.json`;
    const path = `${BACKUP_DIR}/${fileName}`;

    await CloudStore.writeFile(path, backupJson, { override: true });

    onProgress?.({ phase: 'uploading', progress: 0.9 });

    // 3. Log backup record
    await backupApi.log({
      sizeBytes: backupJson.length,
      platform: 'icloud',
      status: 'completed',
    });

    onProgress?.({ phase: 'uploading', progress: 1.0 });
  } catch (error) {
    try {
      await backupApi.log({
        sizeBytes: 0,
        platform: 'icloud',
        status: 'failed',
      });
    } catch {
      // ignore logging error
    }
    throw error;
  }
}

async function icloudRestore(
  onProgress?: ProgressCallback,
): Promise<BackupBundle | null> {
  const CloudStore = (await import('react-native-cloud-store')).default;
  const BACKUP_DIR = 'Documents/Chatat';

  const isAvailable = await CloudStore.isICloudAvailable();
  if (!isAvailable) {
    throw new Error('iCloud is not available');
  }

  // List backups
  onProgress?.({ phase: 'downloading', progress: 0.2 });
  const exists = await CloudStore.exist(BACKUP_DIR);
  if (!exists) return null;

  const files: string[] = await CloudStore.readDir(BACKUP_DIR);
  const jsonFiles = files.filter((f) => f.endsWith('.json')).sort().reverse();
  if (jsonFiles.length === 0) return null;

  // Read latest
  onProgress?.({ phase: 'downloading', progress: 0.5 });
  const latestPath = `${BACKUP_DIR}/${jsonFiles[0]}`;
  const content = await CloudStore.readFile(latestPath);
  const bundle = JSON.parse(content) as BackupBundle;

  // Import to server
  onProgress?.({ phase: 'importing', progress: 0.7 });
  await backupApi.import(bundle);

  onProgress?.({ phase: 'importing', progress: 1.0 });
  return bundle;
}

async function icloudListBackups(): Promise<BackupFile[]> {
  try {
    const CloudStore = (await import('react-native-cloud-store')).default;
    const BACKUP_DIR = 'Documents/Chatat';

    const isAvailable = await CloudStore.isICloudAvailable();
    if (!isAvailable) return [];

    const exists = await CloudStore.exist(BACKUP_DIR);
    if (!exists) return [];

    const files: string[] = await CloudStore.readDir(BACKUP_DIR);
    return files
      .filter((f) => f.endsWith('.json'))
      .sort()
      .reverse()
      .map((f) => ({
        id: f,
        name: f,
        sizeBytes: 0,
        createdAt: parseDateFromFilename(f),
      }));
  } catch {
    return [];
  }
}

function parseDateFromFilename(filename: string): string {
  // Extract date from "chatat-2025-01-15.json"
  const match = filename.match(/chatat-(\d{4}-\d{2}-\d{2})/);
  if (match && match[1]) {
    return new Date(match[1]).toISOString();
  }
  return new Date().toISOString();
}

async function icloudIsAvailable(): Promise<boolean> {
  try {
    const CloudStore = (await import('react-native-cloud-store')).default;
    return await CloudStore.isICloudAvailable();
  } catch {
    return false;
  }
}

// --- Public API ---

export function getPlatform(): BackupPlatform {
  return Platform.OS === 'ios' ? 'icloud' : 'google_drive';
}

export async function performBackup(onProgress?: ProgressCallback): Promise<void> {
  if (Platform.OS === 'ios') {
    return icloudBackup(onProgress);
  }
  return googleDriveBackup(onProgress);
}

export async function performRestore(
  onProgress?: ProgressCallback,
): Promise<BackupBundle | null> {
  if (Platform.OS === 'ios') {
    return icloudRestore(onProgress);
  }
  return googleDriveRestore(onProgress);
}

export async function listCloudBackups(): Promise<BackupFile[]> {
  if (Platform.OS === 'ios') {
    return icloudListBackups();
  }
  return googleDriveListBackups();
}

export async function isCloudAvailable(): Promise<boolean> {
  if (Platform.OS === 'ios') {
    return icloudIsAvailable();
  }
  // Google Drive is always available if Play Services exist
  try {
    const { GoogleSignin } = await import(
      '@react-native-google-signin/google-signin'
    );
    await GoogleSignin.hasPlayServices();
    return true;
  } catch {
    return false;
  }
}
