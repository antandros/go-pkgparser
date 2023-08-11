package pkgparser

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/antandros/go-pkgparser/model"
)

type FieldItem struct {
	Type         reflect.Type
	FindWildcard bool
	Name         string
}
type Converter struct {
	Function  func(string, reflect.Type) (interface{}, error)
	StartFrom string
}
type Parser struct {
	Model      any
	baseStruct map[string]FieldItem
	Packages   []interface{}
}

func (p *Parser) findKey(key string) (*FieldItem, error) {
	itemInfo, ok := p.baseStruct[key]
	if !ok {
		for relKey, relVal := range p.baseStruct {
			if strings.EqualFold(relKey, key) {
				itemInfo = relVal
				ok = true
			}
			if relVal.FindWildcard {
				if len(key) > len(relKey) {
					if strings.EqualFold(key[:len(relKey)], relKey) {
						itemInfo = relVal
						ok = true
					}
				}

			}
		}
	}
	if !ok {
		return nil, fmt.Errorf("%s key not found in struct", key)

	}
	return &itemInfo, nil
}
func (p *Parser) convertContact(s string, targetType reflect.Type) (interface{}, error) {
	r := regexp.MustCompile(`^(.*) <(https?://.+|.+@.+)>$`)
	name := ""
	address := ""
	addressType := ""
	matches := r.FindStringSubmatch(s)
	if len(matches) == 3 {
		name = matches[1]
		address = matches[2]

		if strings.Contains(address, "@") {
			addressType = "email"
		} else {
			addressType = "website"
		}

	}
	return model.PackageContact{
		Contact: address,
		Name:    name,
		Type:    addressType,
	}, nil

}
func (p *Parser) convertStringToFloat(s string, targetType reflect.Type) (interface{}, error) {
	if targetType == nil {
		return s, errors.New("tragetType can not be null")
	}
	typeStr := targetType.Kind().String()

	typeBase, err := strconv.Atoi(typeStr[len(typeStr)-3:])
	if err != nil {
		return strconv.ParseFloat(s, 32)
	}
	return strconv.ParseFloat(s, typeBase)

}
func (p *Parser) convertStringToInt(s string, targetType reflect.Type) (interface{}, error) {
	var err error
	if targetType == nil {
		return s, errors.New("tragetType can not be null")
	}
	typeStr := targetType.Kind().String()
	typeStrSubType := typeStr[:3]
	typeBase := 0

	if strings.EqualFold(typeStrSubType, "int") {
		typeBase, err = strconv.Atoi(typeStr[len(typeStr)-3:])
		if err != nil {
			typeBase = 0
		}
		if typeBase == 0 {
			return strconv.Atoi(s)
		} else {
			return strconv.ParseInt(s, 10, typeBase)
		}
	}
	if strings.EqualFold(typeStrSubType, "uint") {
		typeBase, err = strconv.Atoi(typeStr[len(typeStr)-4:])
		if err != nil {
			return nil, err
		}
		if typeBase == 0 {
			return strconv.Atoi(s)
		} else {
			return strconv.ParseUint(s, 10, typeBase)
		}
	}
	return s, errors.New("type not found")
}
func (p *Parser) convertStringToType(s string, targetType reflect.Type) (interface{}, error) {
	var autoConverts []Converter
	autoConverts = append(autoConverts, Converter{
		StartFrom: "int",
		Function:  p.convertStringToInt,
	})
	autoConverts = append(autoConverts, Converter{
		StartFrom: "uint",
		Function:  p.convertStringToInt,
	})
	autoConverts = append(autoConverts, Converter{
		StartFrom: "PackageContact",
		Function:  p.convertContact,
	})

	autoConverts = append(autoConverts, Converter{
		StartFrom: "float",
		Function:  p.convertStringToFloat,
	})

	for _, converter := range autoConverts {
		if len(targetType.Kind().String()) >= len(converter.StartFrom) {
			typeStr := targetType.Kind().String()[:len(converter.StartFrom)]
			if strings.EqualFold(converter.StartFrom, typeStr) {
				return converter.Function(s, targetType)
			}
		}
		if strings.EqualFold(targetType.Name(), converter.StartFrom) {
			return converter.Function(s, targetType)
		}
	}

	if strings.EqualFold(targetType.Name(), "string") {
		return s, nil
	}
	if strings.EqualFold(targetType.String(), "[]string") {
		return strings.Split(s, " "), nil
	}

	return nil, fmt.Errorf("unsupported type: %s", targetType)

}
func (p *Parser) CreateModel() interface{} {
	modelType := reflect.TypeOf(p.Model)
	newObj := reflect.New(modelType).Interface()
	return newObj
}
func (p *Parser) SetValue(key string, val string, model interface{}) (interface{}, error) {

	itemInfo, err := p.findKey(key)
	others := false
	if err != nil {
		itemInfo, err = p.findKey("ParserOtherAppends")

		if err != nil {
			return model, err
		}
		others = true
	}

	modelValue := reflect.ValueOf(model).Elem()

	nameField := modelValue.FieldByName(itemInfo.Name)
	if nameField.IsValid() && nameField.CanSet() {
		if others {
			if strings.EqualFold(nameField.Type().String(), "map[string]interface {}") {
				mapItem := make(map[string]interface{})
				iter := nameField.MapRange()
				for iter.Next() {
					mapItem[iter.Key().String()] = iter.Value()
				}
				mapItem[key] = val
				nameField.Set(reflect.ValueOf(mapItem))
			}
		} else {

			data, err := p.convertStringToType(val, nameField.Type())
			if err != nil {
				return model, err
			}
			nameField.Set(reflect.ValueOf(data))
		}

		//nameField.Set()
	}
	return model, nil
}

func (p *Parser) StructParse() error {
	baseStruct := make(map[string]FieldItem)
	model := p.Model
	typeItem := reflect.TypeOf(model)
	for i := 0; i < typeItem.NumField(); i++ {
		fld := typeItem.Field(i)
		value, ok := fld.Tag.Lookup("parser")
		if !ok {
			value = fld.Name
		}
		valEn := strings.Split(value, ",")
		if len(valEn) > 1 {
			if strings.EqualFold("others", valEn[1]) {
				value = "ParserOtherAppends"
			} else {
				value = valEn[0]
			}

		}
		findWildcard := false
		if strings.EqualFold(value[len(value)-1:], "*") {
			findWildcard = true
			value = value[:len(value)-1]
		}

		baseStruct[value] = FieldItem{
			Type:         fld.Type,
			Name:         fld.Name,
			FindWildcard: findWildcard,
		}
	}
	if len(baseStruct) == 0 {
		return errors.New("given struct not enouth arguments")
	}
	p.baseStruct = baseStruct
	return nil
}
