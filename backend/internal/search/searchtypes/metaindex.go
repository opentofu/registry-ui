package searchtypes

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

func NewMetaIndex() MetaIndex {
	return MetaIndex{
		Items:         map[IndexID]IndexItem{},
		Deletions:     map[IndexID]time.Time{},
		itemsByParent: map[IndexID]map[IndexID]struct{}{},
		lock:          &sync.Mutex{},
	}
}

type MetaIndex struct {
	Items map[IndexID]IndexItem `json:"items"`
	// Deletions are items that have been removed from the underlying data structure. These deletions will be kept in
	// the metaindex for a period of 30 days to ensure that any search indexes can be updated incrementally.
	Deletions map[IndexID]time.Time `json:"deletions"`

	itemsByParent map[IndexID]map[IndexID]struct{}
	lock          *sync.Mutex
}

func (m *MetaIndex) UnmarshalJSON(data []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	type metaIndex struct {
		Items     map[IndexID]IndexItem `json:"items"`
		Deletions map[IndexID]time.Time `json:"deletions"`
	}

	unmarshalled := metaIndex{
		Items:     map[IndexID]IndexItem{},
		Deletions: map[IndexID]time.Time{},
	}
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		return err
	}
	m.Items = map[IndexID]IndexItem{}
	m.Deletions = map[IndexID]time.Time{}

	m.itemsByParent = map[IndexID]map[IndexID]struct{}{}
	for _, item := range unmarshalled.Items {
		m.addItem(item)
	}
	for i, t := range unmarshalled.Deletions {
		// Only load the deletion if 30 days have not passed since.
		if t.Add(time.Hour * 24 * 30).After(time.Now()) {
			m.Deletions[i] = t
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

	m.addItem(i)
	return nil
}

func (m *MetaIndex) addItem(i IndexItem) {
	if existingItem, ok := m.Items[i.ID]; ok {
		if existingItem.Equals(i) {
			return
		}
	}

	delete(m.Deletions, i.ID)

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

}

func (m *MetaIndex) RemoveItem(_ context.Context, id IndexID) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.removeItem(id)
}

func (m *MetaIndex) removeItem(id IndexID) error {
	if _, ok := m.Deletions[id]; ok {
		return nil
	}

	for id := range m.itemsByParent[id] {
		if err := m.removeItem(id); err != nil {
			return fmt.Errorf("failed to remove subitem %s (%w)", id, err)
		}
	}
	m.Deletions[id] = time.Now()
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
		if err := m.removeItem(id); err != nil {
			return fmt.Errorf("failed to remove item %s (%w)", id, err)
		}
	}
	return nil
}
