package global

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"reflect"
)

var Db *gorm.DB

func InitPredicate(ctx *gin.Context, input interface{}) bool {
	err := ctx.ShouldBindJSON(input)
	if !InputError(ctx, err) {
		return false
	}
	empty := StructEmpty(input)
	if !empty {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "参数不能为空",
		})
		return false
	}
	return true
}
func InputError(ctx *gin.Context, err error) bool {
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    nil,
			"message": "请输入正确的参数",
		})
		return false
	}
	return true
}

// IsStructEmpty 检查结构体中的所有字段是否为空
func StructEmpty(s interface{}) bool {
	// 确保输入是一个结构体
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return false
	}
	val = val.Elem()
	// 遍历结构体的所有字段
	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		switch fieldValue.Kind() {
		case reflect.String:
			if fieldValue.String() == "" {
				return false
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fieldValue.Int() == 0 {
				return false
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if fieldValue.Uint() == 0 {
				return false
			}
		case reflect.Float32, reflect.Float64:
			if fieldValue.Float() == 0 {
				return false
			}
		case reflect.Bool:
			if fieldValue.Bool() {
				return false
			}
		case reflect.Slice, reflect.Array:
			if fieldValue.Len() == 0 {
				return false
			}
		case reflect.Map:
			if fieldValue.Len() == 0 {
				return false
			}
		case reflect.Ptr, reflect.Interface:
			if !fieldValue.IsNil() {
				return false
			}
		default:
			// 对于其他类型，暂时认为为空
			continue
		}
	}

	// 所有字段都为空
	return true
}
