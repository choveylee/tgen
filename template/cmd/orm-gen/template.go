package main

import (
	"gorm.io/gen"
)

// TemplateMethod defines example query methods for generated GORM interfaces.
// The method declarations follow the reference pattern described at
// https://juejin.cn/post/7133150674400837668.
type TemplateMethod interface {
	// WHERE("project_id = @projectId")
	FindByProjectId(projectId uint64) (gen.T, error)

	// WHERE("engineering_id = @engineeringId")
	FindByEngineeringId(engineeringId uint64) (gen.T, error)

	// WHERE("product = @product")
	FindByProduct(product string) (gen.T, error)
}
