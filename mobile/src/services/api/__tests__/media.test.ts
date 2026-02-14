jest.mock('../client', () => ({
  __esModule: true,
  default: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}));

import apiClient from '../client';
import { mediaApi } from '../media';

const mock = apiClient as jest.Mocked<typeof apiClient>;

beforeEach(() => jest.clearAllMocks());

describe('mediaApi', () => {
  it('upload calls post with FormData', async () => {
    mock.post.mockResolvedValue({ data: {} });
    await mediaApi.upload('file://photo.jpg', 'photo.jpg', 'image/jpeg');
    expect(mock.post).toHaveBeenCalledWith(
      '/media/upload',
      expect.any(FormData),
      expect.objectContaining({
        headers: { 'Content-Type': 'multipart/form-data' },
      }),
    );
  });

  it('upload with options', async () => {
    mock.post.mockResolvedValue({ data: {} });
    const progress = jest.fn();
    await mediaApi.upload('file://photo.jpg', 'photo.jpg', 'image/jpeg', {
      contextType: 'chat',
      contextId: 'c1',
      onProgress: progress,
    });
    expect(mock.post).toHaveBeenCalled();
  });

  it('getById calls get', async () => {
    mock.get.mockResolvedValue({ data: {} });
    await mediaApi.getById('m1');
    expect(mock.get).toHaveBeenCalledWith('/media/m1');
  });

  it('getDownloadURL calls get with options', async () => {
    mock.get.mockResolvedValue({ data: 'url' });
    await mediaApi.getDownloadURL('m1');
    expect(mock.get).toHaveBeenCalledWith('/media/m1/download', expect.objectContaining({
      maxRedirects: 0,
    }));
  });

  it('delete calls delete', async () => {
    mock.delete.mockResolvedValue({});
    await mediaApi.delete('m1');
    expect(mock.delete).toHaveBeenCalledWith('/media/m1');
  });
});
