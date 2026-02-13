// useSearch — debounced search hook
import { useState, useEffect, useRef, useCallback } from 'react';
import { searchApi } from '@/services/api/search';
import type {
  MessageSearchResult,
  DocumentSearchResult,
  ContactSearchResult,
  EntitySearchResult,
} from '@/services/api/search';

export type SearchTab = 'all' | 'messages' | 'documents' | 'contacts' | 'entities';

interface SearchAllResults {
  messages: MessageSearchResult[];
  documents: DocumentSearchResult[];
  contacts: ContactSearchResult[];
  entities: EntitySearchResult[];
}

const emptyResults: SearchAllResults = {
  messages: [],
  documents: [],
  contacts: [],
  entities: [],
};

export function useSearch(query: string, tab: SearchTab, debounceMs = 300) {
  const [allResults, setAllResults] = useState<SearchAllResults>(emptyResults);
  const [messages, setMessages] = useState<MessageSearchResult[]>([]);
  const [documents, setDocuments] = useState<DocumentSearchResult[]>([]);
  const [contacts, setContacts] = useState<ContactSearchResult[]>([]);
  const [entities, setEntities] = useState<EntitySearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const search = useCallback(async (q: string, t: SearchTab) => {
    if (q.length < 2) {
      setAllResults(emptyResults);
      setMessages([]);
      setDocuments([]);
      setContacts([]);
      setEntities([]);
      setError(null);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      switch (t) {
        case 'all': {
          const res = await searchApi.searchAll(q, 5);
          setAllResults(res.data.data);
          break;
        }
        case 'messages': {
          const res = await searchApi.searchMessages(q);
          setMessages(res.data.data);
          break;
        }
        case 'documents': {
          const res = await searchApi.searchDocuments(q);
          setDocuments(res.data.data);
          break;
        }
        case 'contacts': {
          const res = await searchApi.searchContacts(q);
          setContacts(res.data.data);
          break;
        }
        case 'entities': {
          const res = await searchApi.searchEntities(q);
          setEntities(res.data.data);
          break;
        }
      }
    } catch {
      setError('Gagal memuat hasil pencarian');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }

    timerRef.current = setTimeout(() => {
      search(query, tab);
    }, debounceMs);

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, [query, tab, debounceMs, search]);

  return {
    allResults,
    messages,
    documents,
    contacts,
    entities,
    isLoading,
    error,
  };
}
