package engine

import "github.com/zheskett/go-voxel/internal/common"

// The global reference frame is exactly what you would expect
func (engine *Engine) BasisFrame() common.ReferenceFramef {
	return common.GlobalFramef()
}

func (engine *Engine) BasisFramei() common.ReferenceFramei {
	return common.GlobalFramei()
}
