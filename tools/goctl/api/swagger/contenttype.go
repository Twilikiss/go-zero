package swagger

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

// 新增：检查是否包含文件类型
func typeContainsFileType(ctx Context, structType spec.DefineStruct) bool {
	members := expandMembers(ctx, structType)
	for _, member := range members {
		tags, _ := spec.Parse(member.Tag)
		if formTag, err := tags.Get(tagForm); err == nil {
			// 检查是否有type=file标记
			for _, option := range formTag.Options {
				if strings.HasPrefix(option, "type=") {
					typeValue := strings.TrimPrefix(option, "type=")
					if typeValue == "file" {
						return true
					}
				}
			}

			// 检查是否是[]byte类型
			if arrayType, ok := member.Type.(spec.ArrayType); ok {
				if primitiveType, ok := arrayType.Value.(spec.PrimitiveType); ok {
					if primitiveType.RawName == "byte" {
						return true
					}
				}
			}
		}
	}
	return false
}

func consumesFromTypeOrDef(ctx Context, method string, tp spec.Type) []string {
	if strings.EqualFold(method, http.MethodGet) {
		return []string{}
	}
	if tp == nil {
		return []string{}
	}
	structType, ok := tp.(spec.DefineStruct)
	if !ok {
		return []string{}
	}

	// 检查是否包含文件类型
	if typeContainsFileType(ctx, structType) {
		return []string{multipartForm}
	}

	if typeContainsTag(ctx, structType, tagJson) {
		return []string{applicationJson}
	}
	return []string{applicationForm}
}
