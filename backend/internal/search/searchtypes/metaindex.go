package searchtypes

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

func NewMetaIndex() MetaIndex {
	return MetaIndex{
		Items:         map[IndexID]IndexItem{},
		itemsByParent: map[IndexID]map[IndexID]struct{}{},
		lock:          &sync.Mutex{},
	}
}

type MetaIndex struct {
	Items map[IndexID]IndexItem `json:"items"`

	itemsByParent map[IndexID]map[IndexID]struct{}
	lock          *sync.Mutex
}

func (m *MetaIndex) UnmarshalJSON(data []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	type metaIndex struct {
		Items map[IndexID]IndexItem `json:"items"`
	}

	unmarshalled := metaIndex{
		Items: nil,
	}
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		return err
	}
	m.Items = map[IndexID]IndexItem{}
	m.itemsByParent = map[IndexID]map[IndexID]struct{}{}
	for _, item := range unmarshalled.Items {
		if err := m.addItem(item); err != nil {
			return err
		}
	}
	for _, item := range m.Items {
		if item.ParentID != "" {
			if _, ok := m.Items[item.ParentID]; !ok {
				return fmt.Errorf("metaindex corrupt, item with ID %s references parent ID %s, which does not exist", item.ID, item.ParentID)
			}
		}
	}
	return nil
}

func (m *MetaIndex) AddItem(_ context.Context, i IndexItem) error {
	if err := i.Validate(); err != nil {
		return err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if i.ParentID != "" {
		if _, ok := m.Items[i.ParentID]; !ok {
			return fmt.Errorf("parent ID not found in index: %s", i.ParentID)
		}
	}

	return m.addItem(i)
}

func (m *MetaIndex) addItem(i IndexItem) error {
	m.Items[i.ID] = i
	if _, ok := m.itemsByParent[i.ID]; !ok {
		m.itemsByParent[i.ID] = map[IndexID]struct{}{}
	}

	if i.ParentID != "" {
		if _, ok := m.itemsByParent[i.ParentID]; !ok {
			m.itemsByParent[i.ParentID] = map[IndexID]struct{}{}
		}
		m.itemsByParent[i.ParentID][i.ID] = struct{}{}
	}
	return nil
}

func (m *MetaIndex) RemoveItem(_ context.Context, id IndexID) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.removeItem(id)
}

func (m *MetaIndex) removeItem(id IndexID) error {
	for id := range m.itemsByParent[id] {
		if err := m.removeItem(id); err != nil {
			return fmt.Errorf("failed to remove subitem %s (%w)", id, err)
		}
	}
	delete(m.itemsByParent, id)
	return nil
}

func (m *MetaIndex) RemoveVersionItems(_ context.Context, itemType IndexType, addr string, version string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var removeQueue []IndexID
	for id, item := range m.Items {
		if item.Type == itemType && item.Addr == addr && item.Version == version {
			removeQueue = append(removeQueue, id)
		}
	}
	for _, id := range removeQueue {
		delete(m.Items, id)
	}
	return nil
}
