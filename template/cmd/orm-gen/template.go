/**
 * @Author: lidonglin
 * @Description:
 * @File:  template.go
 * @Version: 1.0.0
 * @Date: 2023/12/26 14:40
 */

package main

import (
	"gorm.io/gen"
)

// demo: https://juejin.cn/post/7133150674400837668

type TemplateMethod interface {
	// WHERE("project_id = @projectId")
	FindByProjectId(projectId uint64) (gen.T, error)

	// WHERE("engineering_id = @engineeringId")
	FindByEngineeringId(engineeringId uint64) (gen.T, error)

	// WHERE("product = @product")
	FindByProduct(product string) (gen.T, error)
}
