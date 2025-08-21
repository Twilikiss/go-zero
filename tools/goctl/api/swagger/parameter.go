package swagger

import (
	"net/http"
	"strings"

	"github.com/go-openapi/spec"
	apiSpec "github.com/zeromicro/go-zero/tools/goctl/api/spec"
)

func isPostJson(ctx Context, method string, tp apiSpec.Type) (string, bool) {
	//if !strings.EqualFold(method, http.MethodPost) {
	//	return "", false
	//}

	// 【临时】支持所有应该有 request body 的方法（目前支持Post）
	if !strings.EqualFold(method, http.MethodGet) &&
		!strings.EqualFold(method, http.MethodPost) &&
		!strings.EqualFold(method, http.MethodPut) &&
		!strings.EqualFold(method, http.MethodPatch) &&
		!strings.EqualFold(method, http.MethodDelete) {
		return "", false
	}

	structType, ok := tp.(apiSpec.DefineStruct)
	if !ok {
		return "", false
	}
	var isPostJson bool
	rangeMemberAndDo(ctx, structType, func(tag *apiSpec.Tags, required bool, member apiSpec.Member) {
		jsonTag, _ := tag.Get(tagJson)
		if !isPostJson {
			isPostJson = jsonTag != nil
		}
	})
	return structType.RawName, isPostJson
}

// 新增：检查是否是文件类型
func isFileType(formTag *apiSpec.Tag, member apiSpec.Member) bool {
	// 检查form tag的options中是否包含 type=file
	for _, option := range formTag.Options {
		if strings.HasPrefix(option, "type=") {
			typeValue := strings.TrimPrefix(option, "type=")
			if typeValue == "file" {
				return true
			}
		}
	}

	// 检查是否是[]byte类型
	if arrayType, ok := member.Type.(apiSpec.ArrayType); ok {
		if primitiveType, ok := arrayType.Value.(apiSpec.PrimitiveType); ok {
			if primitiveType.RawName == "byte" {
				return true
			}
		}
	}

	return false
}

func parametersFromType(ctx Context, method string, tp apiSpec.Type) []spec.Parameter {
	if tp == nil {
		return []spec.Parameter{}
	}
	structType, ok := tp.(apiSpec.DefineStruct)
	if !ok {
		return []spec.Parameter{}
	}

	var (
		resp           []spec.Parameter
		properties     = map[string]spec.Schema{}
		requiredFields []string
	)
	rangeMemberAndDo(ctx, structType, func(tag *apiSpec.Tags, required bool, member apiSpec.Member) {
		headerTag, _ := tag.Get(tagHeader)
		hasHeader := headerTag != nil

		pathParameterTag, _ := tag.Get(tagPath)
		hasPathParameter := pathParameterTag != nil

		formTag, _ := tag.Get(tagForm)
		hasForm := formTag != nil

		jsonTag, _ := tag.Get(tagJson)
		hasJson := jsonTag != nil
		if hasHeader {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(headerTag.Options)
			// 🎯 新增：为header参数添加example支持
			headerParam := spec.Parameter{
				CommonValidations: spec.CommonValidations{
					Maximum:          maximum,
					ExclusiveMaximum: exclusiveMaximum,
					Minimum:          minimum,
					ExclusiveMinimum: exclusiveMinimum,
					Enum:             enumsValueFromOptions(headerTag.Options),
				},
				SimpleSchema: spec.SimpleSchema{
					Type:    sampleTypeFromGoType(ctx, member.Type),
					Default: defValueFromOptions(ctx, headerTag.Options, member.Type),
					Items:   sampleItemsFromGoType(ctx, member.Type),
				},
				ParamProps: spec.ParamProps{
					In:          paramsInHeader,
					Name:        headerTag.Name,
					Description: formatComment(member.Comment),
					Required:    required,
				},
			}

			// 🎯 为header参数添加example
			if example := exampleValueFromOptions(ctx, headerTag.Options, member.Type); example != nil {
				// 使用VendorExtensible添加example
				if headerParam.VendorExtensible.Extensions == nil {
					headerParam.VendorExtensible.Extensions = make(map[string]interface{})
				}
				headerParam.VendorExtensible.Extensions["x-example"] = example
			}

			resp = append(resp, headerParam)
		}
		if hasPathParameter {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(pathParameterTag.Options)
			// 🎯 新增：为path参数添加example支持
			pathParam := spec.Parameter{
				CommonValidations: spec.CommonValidations{
					Maximum:          maximum,
					ExclusiveMaximum: exclusiveMaximum,
					Minimum:          minimum,
					ExclusiveMinimum: exclusiveMinimum,
					Enum:             enumsValueFromOptions(pathParameterTag.Options),
				},
				SimpleSchema: spec.SimpleSchema{
					Type:    sampleTypeFromGoType(ctx, member.Type),
					Default: defValueFromOptions(ctx, pathParameterTag.Options, member.Type),
					Items:   sampleItemsFromGoType(ctx, member.Type),
				},
				ParamProps: spec.ParamProps{
					In:          paramsInPath,
					Name:        pathParameterTag.Name,
					Description: formatComment(member.Comment),
					Required:    required,
				},
			}

			// 🎯 为path参数添加example
			if example := exampleValueFromOptions(ctx, pathParameterTag.Options, member.Type); example != nil {
				// 使用VendorExtensible添加example
				if pathParam.VendorExtensible.Extensions == nil {
					pathParam.VendorExtensible.Extensions = make(map[string]interface{})
				}
				pathParam.VendorExtensible.Extensions["x-example"] = example
			}

			resp = append(resp, pathParam)
		}
		if hasForm {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(formTag.Options)
			if strings.EqualFold(method, http.MethodGet) {
				// 🎯 GET方法的form参数（作为query参数）
				queryParam := spec.Parameter{
					CommonValidations: spec.CommonValidations{
						Maximum:          maximum,
						ExclusiveMaximum: exclusiveMaximum,
						Minimum:          minimum,
						ExclusiveMinimum: exclusiveMinimum,
						Enum:             enumsValueFromOptions(formTag.Options),
					},
					SimpleSchema: spec.SimpleSchema{
						Type:    sampleTypeFromGoType(ctx, member.Type),
						Default: defValueFromOptions(ctx, formTag.Options, member.Type),
						Items:   sampleItemsFromGoType(ctx, member.Type),
					},
					ParamProps: spec.ParamProps{
						In:              paramsInQuery,
						Name:            formTag.Name,
						Description:     formatComment(member.Comment),
						Required:        required,
						AllowEmptyValue: !required,
					},
				}

				// 🎯 为query参数添加example
				if example := exampleValueFromOptions(ctx, formTag.Options, member.Type); example != nil {
					// 使用VendorExtensible添加example
					if queryParam.VendorExtensible.Extensions == nil {
						queryParam.VendorExtensible.Extensions = make(map[string]interface{})
					}
					queryParam.VendorExtensible.Extensions["x-example"] = example
				}

				resp = append(resp, queryParam)
			} else {
				// POST等方法的form参数
				// 检查是否是文件类型
				if isFileType(formTag, member) {
					// 处理文件类型参数
					fileParam := spec.Parameter{
						CommonValidations: spec.CommonValidations{
							Maximum:          maximum,
							ExclusiveMaximum: exclusiveMaximum,
							Minimum:          minimum,
							ExclusiveMinimum: exclusiveMinimum,
							Enum:             enumsValueFromOptions(formTag.Options),
						},
						SimpleSchema: spec.SimpleSchema{
							Type:   "file",
							Format: "binary",
						},
						ParamProps: spec.ParamProps{
							In:              paramsInForm,
							Name:            formTag.Name,
							Description:     formatComment(member.Comment),
							Required:        required,
							AllowEmptyValue: !required,
						},
					}

					// 🎯 文件类型通常不需要example，但如果有的话也添加
					if example := exampleValueFromOptions(ctx, formTag.Options, member.Type); example != nil {
						if fileParam.VendorExtensible.Extensions == nil {
							fileParam.VendorExtensible.Extensions = make(map[string]interface{})
						}
						fileParam.VendorExtensible.Extensions["x-example"] = example
					}

					resp = append(resp, fileParam)
				} else {
					// 处理普通form参数
					formParam := spec.Parameter{
						CommonValidations: spec.CommonValidations{
							Maximum:          maximum,
							ExclusiveMaximum: exclusiveMaximum,
							Minimum:          minimum,
							ExclusiveMinimum: exclusiveMinimum,
							Enum:             enumsValueFromOptions(formTag.Options),
						},
						SimpleSchema: spec.SimpleSchema{
							Type:    sampleTypeFromGoType(ctx, member.Type),
							Default: defValueFromOptions(ctx, formTag.Options, member.Type),
							Items:   sampleItemsFromGoType(ctx, member.Type),
						},
						ParamProps: spec.ParamProps{
							In:              paramsInForm,
							Name:            formTag.Name,
							Description:     formatComment(member.Comment),
							Required:        required,
							AllowEmptyValue: !required,
						},
					}

					// 🎯 为form参数添加example
					if example := exampleValueFromOptions(ctx, formTag.Options, member.Type); example != nil {
						// 使用VendorExtensible添加example
						if formParam.VendorExtensible.Extensions == nil {
							formParam.VendorExtensible.Extensions = make(map[string]interface{})
						}
						formParam.VendorExtensible.Extensions["x-example"] = example
					}

					resp = append(resp, formParam)
				}
			}

		}
		if hasJson {
			minimum, maximum, exclusiveMinimum, exclusiveMaximum := rangeValueFromOptions(jsonTag.Options)
			if required {
				requiredFields = append(requiredFields, jsonTag.Name)
			}
			var schema = spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					Example: exampleValueFromOptions(ctx, jsonTag.Options, member.Type),
				},
				SchemaProps: spec.SchemaProps{
					Description:          formatComment(member.Comment),
					Type:                 typeFromGoType(ctx, member.Type),
					Default:              defValueFromOptions(ctx, jsonTag.Options, member.Type),
					Maximum:              maximum,
					ExclusiveMaximum:     exclusiveMaximum,
					Minimum:              minimum,
					ExclusiveMinimum:     exclusiveMinimum,
					Enum:                 enumsValueFromOptions(jsonTag.Options),
					AdditionalProperties: mapFromGoType(ctx, member.Type),
				},
			}
			switch sampleTypeFromGoType(ctx, member.Type) {
			case swaggerTypeArray:
				// 检查数组元素类型
				if arrayType, ok := member.Type.(apiSpec.ArrayType); ok {
					if defineType, ok := arrayType.Value.(apiSpec.DefineStruct); ok && ctx.UseDefinitions {
						// 如果是定义结构体且UseDefinitions=true，使用$ref引用
						schema.Items = &spec.SchemaOrArray{
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Ref: spec.MustCreateRef("#/definitions/" + defineType.RawName),
								},
							},
						}
					} else {
						// 否则使用普通的itemsFromGoType
						schema.Items = itemsFromGoType(ctx, member.Type)
					}
				} else {
					schema.Items = itemsFromGoType(ctx, member.Type)
				}
			case swaggerTypeObject:
				p, r := propertiesFromType(ctx, member.Type)
				schema.Properties = p
				schema.Required = r
			}
			properties[jsonTag.Name] = schema
		}
	})

	// 处理JSON请求体
	if len(properties) > 0 {
		if ctx.UseDefinitions {
			structName, ok := isPostJson(ctx, method, tp)
			if ok {
				bodyParam := spec.Parameter{
					ParamProps: spec.ParamProps{
						In:       paramsInBody,
						Name:     paramsInBody,
						Required: true,
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: spec.MustCreateRef(getRefName(structName)),
							},
						},
					},
				}
				resp = append(resp, bodyParam)
			}
		} else {
			bodyParam := spec.Parameter{
				ParamProps: spec.ParamProps{
					In:       paramsInBody,
					Name:     paramsInBody,
					Required: true,
					Schema: &spec.Schema{
						SchemaProps: spec.SchemaProps{
							Type:       typeFromGoType(ctx, structType),
							Properties: properties,
							Required:   requiredFields,
						},
					},
				},
			}
			resp = append(resp, bodyParam)
		}
	}
	return resp
}
