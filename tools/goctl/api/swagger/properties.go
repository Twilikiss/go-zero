package swagger

import (
	"github.com/go-openapi/spec"
	apiSpec "github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"strings"
)

func propertiesFromType(ctx Context, tp apiSpec.Type) (spec.SchemaProperties, []string) {
	var (
		properties     = map[string]spec.Schema{}
		requiredFields []string
	)
	switch val := tp.(type) {
	case apiSpec.PointerType:
		return propertiesFromType(ctx, val.Type)
	case apiSpec.ArrayType:
		return propertiesFromType(ctx, val.Value)
	case apiSpec.DefineStruct, apiSpec.NestedStruct:
		rangeMemberAndDo(ctx, val, func(tag *apiSpec.Tags, required bool, member apiSpec.Member) {
			var (
				jsonTagString                      = member.Name
				minimum, maximum                   *float64
				exclusiveMinimum, exclusiveMaximum bool
				example, defaultValue              any
				enum                               []any
			)
			pathTag, _ := tag.Get(tagPath)
			if pathTag != nil {
				return
			}
			formTag, _ := tag.Get(tagForm)
			if formTag != nil {
				return
			}
			headerTag, _ := tag.Get(tagHeader)
			if headerTag != nil {
				return
			}

			jsonTag, _ := tag.Get(tagJson)
			if jsonTag != nil {
				jsonTagString = jsonTag.Name
				minimum, maximum, exclusiveMinimum, exclusiveMaximum = rangeValueFromOptions(jsonTag.Options)
				example = exampleValueFromOptions(ctx, jsonTag.Options, member.Type)
				defaultValue = defValueFromOptions(ctx, jsonTag.Options, member.Type)
				enum = enumsValueFromOptions(jsonTag.Options)

				// 新增：检查json tag中的type=file
				if isFileTypeInJsonTag(jsonTag) {
					schema := spec.Schema{
						SwaggerSchemaProps: spec.SwaggerSchemaProps{
							Example: example,
						},
						SchemaProps: spec.SchemaProps{
							Type:        []string{swaggerTypeString},
							Format:      "binary", // 文件类型使用binary格式
							Description: formatComment(member.Comment),
						},
					}
					properties[jsonTagString] = schema
					if required {
						requiredFields = append(requiredFields, jsonTagString)
					}
					return // 文件类型处理完毕，直接返回
				}
			}

			if required {
				requiredFields = append(requiredFields, jsonTagString)
			}

			schema := spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					Example: example,
				},
				SchemaProps: spec.SchemaProps{
					Description:          formatComment(member.Comment),
					Type:                 typeFromGoType(ctx, member.Type),
					Default:              defaultValue,
					Maximum:              maximum,
					ExclusiveMaximum:     exclusiveMaximum,
					Minimum:              minimum,
					ExclusiveMinimum:     exclusiveMinimum,
					Enum:                 enum,
					AdditionalProperties: mapFromGoType(ctx, member.Type),
				},
			}

			switch sampleTypeFromGoType(ctx, member.Type) {
			case swaggerTypeArray:
				schema.Items = itemsFromGoType(ctx, member.Type)
			case swaggerTypeObject:
				p, r := propertiesFromType(ctx, member.Type)
				schema.Properties = p
				schema.Required = r
			}
			if ctx.UseDefinitions {
				structName, containsStruct := containsStruct(member.Type)
				if containsStruct {
					schema.SchemaProps.Ref = spec.MustCreateRef(getRefName(structName))
				}
			}

			properties[jsonTagString] = schema
		})
	}

	return properties, requiredFields
}

func containsStruct(tp apiSpec.Type) (string, bool) {
	switch val := tp.(type) {
	case apiSpec.PointerType:
		return containsStruct(val.Type)
	case apiSpec.ArrayType:
		return containsStruct(val.Value)
	case apiSpec.DefineStruct:
		return val.RawName, true
	case apiSpec.MapType:
		return containsStruct(val.Value)
	default:
		return "", false
	}
}

func getRefName(typeName string) string {
	return "#/definitions/" + typeName
}

// 新增：检查json tag中是否包含type=file
func isFileTypeInJsonTag(jsonTag *apiSpec.Tag) bool {
	for _, option := range jsonTag.Options {
		if strings.HasPrefix(option, "type=") {
			typeValue := strings.TrimPrefix(option, "type=")
			if typeValue == "file" {
				return true
			}
		}
	}
	return false
}
