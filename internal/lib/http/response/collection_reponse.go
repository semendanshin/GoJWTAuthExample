package response

import "encoding/json"

type CollectionResponse[T Data] struct {
	Response[T]
	TotalCount int `json:"total_count"`
	Offset     int `json:"offset"`
	Limit      int `json:"limit"`
}

func NewCollectionResponse[T Data](data T, count, offset, limit int) *CollectionResponse[T] {
	return &CollectionResponse[T]{
		Response: Response[T]{
			Status: StatusSuccess,
			Data:   data,
		},
		TotalCount: count,
		Offset:     offset,
		Limit:      limit,
	}
}

func (r *CollectionResponse[T]) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
