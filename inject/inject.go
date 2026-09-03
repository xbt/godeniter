// Package inject 提供了纯 Go 标准库（基于 reflect 反射）实现的轻量级依赖注入容器。
// 该模块的设计思想参考了经典框架 Martini (codegangsta/inject)，
// 能够支持在运行时根据函数的入参类型自动解析并注入依赖，使得 HTTP Handler 和中间件的函数签名更加灵活自由。
package inject

import (
	"fmt"
	"reflect"
)

// Injector 定义了依赖注入容器的核心接口。
// 它支持按具体值类型映射、按接口类型映射、结构体依赖注入、函数动态调用以及父子容器层级继承。
type Injector interface {
	// Map 将一个具体的值注入到容器中，类型为其 concrete type（具体类型）。
	// 返回自身以支持链式调用。
	Map(val interface{}) Injector

	// MapTo 将一个值以指定的接口类型（interface type）映射到容器中。
	// ifacePtr 必须是一个指向接口的指针，例如：(*MyInterface)(nil)。
	// 返回自身以支持链式调用。
	MapTo(val interface{}, ifacePtr interface{}) Injector

	// Get 根据反射类型检索容器中注入的值。
	// 如果当前容器未找到，会递归向父容器查找；若均未找到则返回零值 reflect.Value{}。
	Get(t reflect.Type) reflect.Value

	// SetParent 设置父级注入器。
	// 在 Web 框架中常用于：全局 App 容器作为父容器，每个 HTTP Request 创建一个子容器继承全局服务。
	SetParent(parent Injector)

	// Invoke 反射解析传入函数 f 的入参列表，从容器中查找匹配的依赖并自动执行该函数。
	// 返回函数的执行返回值列表；若入参类型无法在容器中解析，则返回 error。
	Invoke(f interface{}) ([]reflect.Value, error)

	// Apply 扫描传入结构体指针的导出字段，对带有 `inject:""` 标签的字段自动从容器中注入依赖。
	Apply(val interface{}) error
}

// injector 是 Injector 接口的默认实现。
type injector struct {
	values map[reflect.Type]reflect.Value // 类型到反射值的映射字典
	parent Injector                       // 父级注入器指针，用于层级查找
}

// New 创建并返回一个新的依赖注入容器实例。
func New() Injector {
	return &injector{
		values: make(map[reflect.Type]reflect.Value),
	}
}

// Map 将一个具体的值注入容器。
// 示例：
//
//	inj := inject.New()
//	inj.Map(&sql.DB{}) // 容器将记录 *sql.DB 类型对应的值
func (inj *injector) Map(val interface{}) Injector {
	inj.values[reflect.TypeOf(val)] = reflect.ValueOf(val)
	return inj
}

// MapTo 将一个值以指定的接口类型映射注入容器。
// 示例：
//
//	type Service interface { Do() }
//	type myService struct{}
//	inj.MapTo(&myService{}, (*Service)(nil))
func (inj *injector) MapTo(val interface{}, ifacePtr interface{}) Injector {
	t := reflect.TypeOf(ifacePtr)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Interface {
		panic(fmt.Sprintf("inject: MapTo 的第二个参数必须是指向接口的指针，收到: %v", t))
	}
	inj.values[t.Elem()] = reflect.ValueOf(val)
	return inj
}

// Get 根据反射类型检索注入的值。
func (inj *injector) Get(t reflect.Type) reflect.Value {
	val, ok := inj.values[t]
	if ok {
		return val
	}
	// 如果当前容器找不到且存在父容器，则委托给父容器查找
	if inj.parent != nil {
		return inj.parent.Get(t)
	}
	return reflect.Value{}
}

// SetParent 设置父级注入器。
func (inj *injector) SetParent(parent Injector) {
	inj.parent = parent
}

// Invoke 解析函数入参依赖并执行函数。
// 示例：
//
//	inj.Map("Hello Godeniter")
//	inj.Invoke(func(msg string) {
//	    fmt.Println(msg) // 输出: Hello Godeniter
//	})
func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(f)
	if t == nil || t.Kind() != reflect.Func {
		return nil, fmt.Errorf("inject: Invoke 期望传入函数，收到: %v", t)
	}

	numIn := t.NumIn()
	inValues := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		inType := t.In(i)
		val := inj.Get(inType)
		if !val.IsValid() {
			return nil, fmt.Errorf("inject: 无法在容器中找到类型为 [%s] 的依赖参数 (函数第 %d 个参数)", inType.String(), i+1)
		}
		inValues[i] = val
	}

	// 反射调用目标函数并返回结果
	return reflect.ValueOf(f).Call(inValues), nil
}

// Apply 为结构体带有 `inject:""` 标签的字段自动注入依赖。
// val 必须是一个非空的结构体指针。
func (inj *injector) Apply(val interface{}) error {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("inject: Apply 期望传入结构体指针，收到: %v", v.Type())
	}

	elem := v.Elem()
	elemType := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := elemType.Field(i)

		// 仅对带有 `inject` tag 且可设置的字段进行注入
		if _, ok := fieldType.Tag.Lookup("inject"); ok && field.CanSet() {
			val := inj.Get(fieldType.Type)
			if !val.IsValid() {
				return fmt.Errorf("inject: 无法为结构体字段 %s 找到类型为 [%s] 的依赖", fieldType.Name, fieldType.Type.String())
			}
			field.Set(val)
		}
	}

	return nil
}
