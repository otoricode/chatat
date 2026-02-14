// @ts-nocheck
jest.mock('@/services/api/entities', () => ({
  entitiesApi: {
    list: jest.fn(),
    listTypes: jest.fn(),
    search: jest.fn(),
    create: jest.fn(),
    update: jest.fn(),
    delete: jest.fn(),
    createFromContact: jest.fn(),
  },
}));

import { useEntityStore } from '../entityStore';
import { entitiesApi } from '@/services/api/entities';

const mockEntitiesApi = entitiesApi as jest.Mocked<typeof entitiesApi>;

const makeEntity = (id: string, name: string) => ({
  id,
  name,
  type: 'person',
  fields: {},
});

beforeEach(() => {
  useEntityStore.setState({
    entities: [],
    types: [],
    isLoading: false,
    total: 0,
    error: null,
  });
  jest.clearAllMocks();
});

describe('entityStore', () => {
  it('starts empty', () => {
    const s = useEntityStore.getState();
    expect(s.entities).toEqual([]);
    expect(s.types).toEqual([]);
  });

  it('fetchEntities success', async () => {
    const entities = [makeEntity('e1', 'Alice'), makeEntity('e2', 'Bob')];
    mockEntitiesApi.list.mockResolvedValue({
      data: { data: { data: entities, total: 2, limit: 20, offset: 0 } },
    } as any);

    await useEntityStore.getState().fetchEntities();

    const s = useEntityStore.getState();
    expect(s.entities).toHaveLength(2);
    expect(s.total).toBe(2);
    expect(s.isLoading).toBe(false);
  });

  it('fetchEntities handles null data', async () => {
    mockEntitiesApi.list.mockResolvedValue({
      data: { data: { data: null, total: 0 } },
    } as any);

    await useEntityStore.getState().fetchEntities();

    expect(useEntityStore.getState().entities).toEqual([]);
  });

  it('fetchEntities error with Error', async () => {
    mockEntitiesApi.list.mockRejectedValue(new Error('Network error'));

    await useEntityStore.getState().fetchEntities();

    expect(useEntityStore.getState().error).toBe('Network error');
    expect(useEntityStore.getState().isLoading).toBe(false);
  });

  it('fetchEntities error with non-Error uses i18n', async () => {
    mockEntitiesApi.list.mockRejectedValue('fail');

    await useEntityStore.getState().fetchEntities();

    // i18n is mocked to return the key
    expect(useEntityStore.getState().error).toBe('entity.loadFailed');
  });

  it('fetchTypes success', async () => {
    mockEntitiesApi.listTypes.mockResolvedValue({
      data: { data: ['person', 'organization'] },
    } as any);

    await useEntityStore.getState().fetchTypes();

    expect(useEntityStore.getState().types).toEqual(['person', 'organization']);
  });

  it('fetchTypes handles null data', async () => {
    mockEntitiesApi.listTypes.mockResolvedValue({
      data: { data: null },
    } as any);

    await useEntityStore.getState().fetchTypes();

    expect(useEntityStore.getState().types).toEqual([]);
  });

  it('fetchTypes error is silent', async () => {
    mockEntitiesApi.listTypes.mockRejectedValue(new Error('fail'));

    await useEntityStore.getState().fetchTypes();

    // No error set
    expect(useEntityStore.getState().error).toBeNull();
  });

  it('searchEntities success', async () => {
    const entities = [makeEntity('e1', 'Alice')];
    mockEntitiesApi.search.mockResolvedValue({
      data: { data: entities },
    } as any);

    const result = await useEntityStore.getState().searchEntities('Alice');

    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Alice');
  });

  it('searchEntities handles null data', async () => {
    mockEntitiesApi.search.mockResolvedValue({
      data: { data: null },
    } as any);

    const result = await useEntityStore.getState().searchEntities('test');

    expect(result).toEqual([]);
  });

  it('searchEntities error returns empty', async () => {
    mockEntitiesApi.search.mockRejectedValue(new Error('fail'));

    const result = await useEntityStore.getState().searchEntities('test');

    expect(result).toEqual([]);
  });

  it('createEntity returns entity', async () => {
    const entity = makeEntity('e1', 'New Entity');
    mockEntitiesApi.create.mockResolvedValue({
      data: { data: entity },
    } as any);

    const result = await useEntityStore.getState().createEntity({
      name: 'New Entity',
      type: 'person',
    });

    expect(result.name).toBe('New Entity');
  });

  it('updateEntity returns updated entity', async () => {
    const entity = makeEntity('e1', 'Updated');
    mockEntitiesApi.update.mockResolvedValue({
      data: { data: entity },
    } as any);

    const result = await useEntityStore.getState().updateEntity('e1', {
      name: 'Updated',
    });

    expect(result.name).toBe('Updated');
  });

  it('deleteEntity removes from list', async () => {
    useEntityStore.setState({ entities: [makeEntity('e1', 'Alice'), makeEntity('e2', 'Bob')] as any });
    mockEntitiesApi.delete.mockResolvedValue({} as any);

    await useEntityStore.getState().deleteEntity('e1');

    const entities = useEntityStore.getState().entities;
    expect(entities).toHaveLength(1);
    expect(entities[0].id).toBe('e2');
  });

  it('createFromContact returns entity', async () => {
    const entity = makeEntity('e1', 'Contact Entity');
    mockEntitiesApi.createFromContact.mockResolvedValue({
      data: { data: entity },
    } as any);

    const result = await useEntityStore.getState().createFromContact('u1');

    expect(result.name).toBe('Contact Entity');
  });

  it('clearError resets error', () => {
    useEntityStore.setState({ error: 'some error' });

    useEntityStore.getState().clearError();

    expect(useEntityStore.getState().error).toBeNull();
  });
});
