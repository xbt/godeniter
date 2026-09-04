package db

import (
	"math"
)

// Pagination 分页元数据信息。
type Pagination struct {
	Total      int64 `json:"total"`       // 总记录条数
	Page       int   `json:"page"`        // 当前页码 (从 1 开始)
	PageSize   int   `json:"page_size"`   // 每页记录数
	TotalPages int   `json:"total_pages"` // 总页数
	HasNext    bool  `json:"has_next"`    // 是否有下一页
	HasPrev    bool  `json:"has_prev"`    // 是否有上一页
}

// Paginate 便捷执行分页查询，自动计算 COUNT(*) 统计总数、执行 OFFSET/LIMIT 查询并扫描至 destSlicePtr。
// 示例：
//
//	var users []models.User
//	pager, err := db.Table("users").Where("status = ?", 1).OrderBy("id DESC").Paginate(1, 10, &users)
func (qb *QueryBuilder) Paginate(page, pageSize int, destSlicePtr any) (*Pagination, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 1. 统计当前条件下的记录总数
	total, err := qb.Count()
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages <= 0 {
		totalPages = 1
	}

	// 2. 设置分页偏移并执行查询
	qb.Limit(pageSize).Offset((page - 1) * pageSize)
	if err := qb.Find(destSlicePtr); err != nil {
		return nil, err
	}

	return &Pagination{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}, nil
}
