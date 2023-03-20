package model

import (
	"github.com/uzziahlin/orm/internal/errs"
	"github.com/uzziahlin/orm/utils"
	"reflect"
	"strings"
	"sync"
)

const (
	tagName = "orm"

	columnTag = "column"
)

type TableNamer interface {
	TableName() string
}

// Registry 元数据注册中心抽象
type Registry interface {
	Register(m any, opts ...Option) (*Model, error)

	Get(m any) (*Model, error)
}

func NewRegistry() Registry {
	return &registry{}
}

// registry 元数据注册中心默认实现
type registry struct {
	metas sync.Map
}

func (r *registry) Register(m any, opts ...Option) (*Model, error) {
	typ := reflect.TypeOf(m)

	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.NewErrUnsupportedType(typ.Name())
	}

	typ = typ.Elem()

	parser := &parser{
		m:   m,
		typ: typ,
	}
	// parse Model
	meta, err := parser.parseModel()

	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		err := opt(meta)

		if err != nil {
			return nil, err
		}

	}

	r.metas.Store(typ, meta)

	return meta, nil
}

func (r *registry) Get(m any) (*Model, error) {
	// 先看一下注册中心存不存在元数据，不存在则注册

	typ := reflect.TypeOf(m)

	val, ok := r.metas.Load(typ.Elem())

	if !ok {
		return r.Register(m)
	}

	return val.(*Model), nil
}

type parser struct {
	m   any
	typ reflect.Type
}

func (p *parser) parseModel() (*Model, error) {

	model, err := p.parseModelInfo()

	if err != nil {
		return nil, err
	}

	// parse FieldMap
	fieldMap := make(map[string]*Field, p.typ.NumField())
	columnMap := make(map[string]*Field, p.typ.NumField())
	for i := 0; i < p.typ.NumField(); i++ {
		f := p.typ.Field(i)

		field, err := p.parseFieldInfo(f)

		if err != nil {
			return nil, err
		}

		fieldMap[f.Name] = field
		columnMap[field.ColName] = field

		model.Fields = append(model.Fields, field)
	}
	model.FieldMap = fieldMap
	model.ColumnMap = columnMap

	return model, nil
}

func (p *parser) parseModelInfo() (*Model, error) {

	var m Model

	namer, ok := p.m.(TableNamer)

	if ok {
		m.TabName = namer.TableName()
	} else {
		m.TabName = utils.CamelToUnderLine(p.typ.Name())
	}
	return &m, nil
}

func (p *parser) parseFieldInfo(fd reflect.StructField) (*Field, error) {
	// 解析tag
	tags, err := p.parseTag(fd.Tag)

	if err != nil {
		return nil, err
	}

	// 获取tag表示列名的表示
	col := tags[columnTag]

	// 如果为“”， 则用默认的
	if col == "" {
		col = utils.CamelToUnderLine(fd.Name)
	}

	return &Field{
		GoName:  fd.Name,
		ColName: col,
		GoType:  fd.Type,
		Offset:  fd.Offset,
	}, nil

}

func (p *parser) parseTag(tag reflect.StructTag) (map[string]string, error) {
	val := tag.Get(tagName)

	if val == "" {
		return nil, nil
	}

	pairs := strings.Split(val, ",")

	res := make(map[string]string, len(pairs))

	for _, pair := range pairs {
		segs := strings.Split(pair, "=")

		if len(segs) != 2 {
			return nil, errs.NewErrTagInvalid(pair)
		}

		res[segs[0]] = segs[1]
	}

	return res, nil

}
