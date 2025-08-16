package dto

type Paginated struct {
	Limit  int `form:"limit" validate:"gte=10,lte=100"`
	Offset int `form:"offset" validate:"gte=0"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination struct {
		Total      int `json:"total"`
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		TotalPages int `json:"total_pages"`
	} `json:"pagination"`
}

func MakePaginatedResponse(data interface{}, total int, offset int, limit int) PaginatedResponse {
	totalPages := total / limit
	page := offset/limit + 1

	if total%limit != 0 {
		totalPages++
	}

	return PaginatedResponse{
		Data: data,
		Pagination: struct {
			Total      int `json:"total"`
			Page       int `json:"page"`
			Limit      int `json:"limit"`
			TotalPages int `json:"total_pages"`
		}{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}
}
