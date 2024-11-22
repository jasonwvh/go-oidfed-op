package storage

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/zachmann/go-oidfed/pkg"
	"github.com/zachmann/go-oidfed/pkg/jwk"
)

// SubordinateInfo holds information about a subordinate for storage
type SubordinateInfo struct {
	JWKS               jwk.JWKS                     `json:"jwks"`
	EntityTypes        []string                     `json:"entity_types"`
	EntityID           string                       `json:"entity_id"`
	Metadata           *pkg.Metadata                `json:"metadata,omitempty"`
	MetadataPolicy     *pkg.MetadataPolicies        `json:"metadata_policy,omitempty"`
	Constraints        *pkg.ConstraintSpecification `json:"constraints,omitempty"`
	CriticalExtensions []string                     `json:"crit,omitempty"`
	MetadataPolicyCrit []pkg.PolicyOperatorName     `json:"metadata_policy_crit,omitempty"`
	TrustMarks         pkg.TrustMarkInfos           `json:"trust_marks,omitempty"`
	Extra              map[string]interface{}       `json:"-"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (info *SubordinateInfo) UnmarshalJSON(src []byte) error {
	type subordinateInfo SubordinateInfo
	ii := subordinateInfo(*info)
	if err := json.Unmarshal(src, &ii); err != nil {
		return err
	}
	*info = SubordinateInfo(ii)
	return nil
}

// UnmarshalMsgpack implements the msgpack.Unmarshaler interface
func (info *SubordinateInfo) UnmarshalMsgpack(src []byte) error {
	type subordinateInfo SubordinateInfo
	ii := subordinateInfo(*info)
	if err := msgpack.Unmarshal(src, &ii); err != nil {
		return err
	}
	*info = SubordinateInfo(ii)
	return nil
}

// SubordinateStorageBackend is an interface to store SubordinateInfo
type SubordinateStorageBackend interface {
	Write(entityID string, info SubordinateInfo) error
	Q() SubordinateStorageQuery
	Delete(entityID string) error
	Load() error
}

// SubordinateStorageQuery is an interface to query SubordinateInfo from storage
type SubordinateStorageQuery interface {
	Subordinate(entityID string) (*SubordinateInfo, error)
	Subordinates() ([]SubordinateInfo, error)
	EntityIDs() ([]string, error)
	AddFilter(filter SubordinateStorageQueryFilter, value any) error
}

// SubordinateStorageQueryFilter is function to filter SubordinateInfo
type SubordinateStorageQueryFilter func(info SubordinateInfo, value any) bool

type simpleSubordinateStorageQuery struct {
	filters []func(info SubordinateInfo) bool
	base    []SubordinateInfo
}

func (q *simpleSubordinateStorageQuery) applyFilter() {
	var filtered []SubordinateInfo
	for _, s := range q.base {
		stillOK := true
		for _, f := range q.filters {
			if !f(s) {
				stillOK = false
				break
			}
		}
		if stillOK {
			filtered = append(filtered, s)
		}
	}
	q.base = filtered
}

// Subordinate implements the SubordinateStorageQuery interface
func (q *simpleSubordinateStorageQuery) Subordinate(entityID string) (*SubordinateInfo, error) {
	for _, i := range q.base {
		if i.EntityID == entityID {
			return &i, nil
		}
	}
	return nil, nil
}

// Subordinates implements the SubordinateStorageQuery interface
func (q *simpleSubordinateStorageQuery) Subordinates() ([]SubordinateInfo, error) {
	q.applyFilter()
	return q.base, nil
}

// EntityIDs implements the SubordinateStorageQuery interface
func (q *simpleSubordinateStorageQuery) EntityIDs() ([]string, error) {
	q.applyFilter()
	ids := make([]string, len(q.base))
	for i, info := range q.base {
		ids[i] = info.EntityID
	}
	return ids, nil
}

// AddFilter implements the SubordinateStorageQuery interface
func (q *simpleSubordinateStorageQuery) AddFilter(
	filter SubordinateStorageQueryFilter,
	value any,
) error {
	q.filters = append(
		q.filters, func(info SubordinateInfo) bool {
			return filter(info, value)
		},
	)
	return nil
}
