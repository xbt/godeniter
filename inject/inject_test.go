// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package inject

import (
	"testing"
)

type DatabaseService interface {
	Query() string
}

type mockDB struct {
	Name string
}

func (m *mockDB) Query() string {
	return "result from " + m.Name
}

type AppConfig struct {
	AppName string
	Port    int
}

type Controller struct {
	DB     DatabaseService `inject:""`
	Config *AppConfig      `inject:""`
	Unused string          // 无 inject 标签，不应被注入
}

func TestInjector_MapAndInvoke(t *testing.T) {
	inj := New()
	cfg := &AppConfig{AppName: "GodeniterApp", Port: 8080}
	inj.Map(cfg)
	inj.Map("custom_secret_key")

	called := false
	results, err := inj.Invoke(func(c *AppConfig, secret string) (string, int) {
		called = true
		if c.AppName != "GodeniterApp" || c.Port != 8080 {
			t.Errorf("参数注入错误: %+v", c)
		}
		if secret != "custom_secret_key" {
			t.Errorf("参数注入错误: %s", secret)
		}
		return c.AppName, c.Port
	})

	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if !called {
		t.Errorf("目标函数未被执行")
	}
	if len(results) != 2 || results[0].String() != "GodeniterApp" || results[1].Int() != 8080 {
		t.Errorf("返回值不匹配: %v", results)
	}
}

func TestInjector_MapTo(t *testing.T) {
	inj := New()
	db := &mockDB{Name: "MySQL"}

	// 映射到接口
	inj.MapTo(db, (*DatabaseService)(nil))

	called := false
	_, err := inj.Invoke(func(s DatabaseService) {
		called = true
		if s.Query() != "result from MySQL" {
			t.Errorf("接口方法调用结果不匹配: %s", s.Query())
		}
	})

	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}
	if !called {
		t.Errorf("目标函数未被执行")
	}
}

func TestInjector_ParentChild(t *testing.T) {
	parent := New()
	parent.Map("parent_value")

	child := New()
	child.SetParent(parent)
	child.Map(12345)

	_, err := child.Invoke(func(s string, i int) {
		if s != "parent_value" || i != 12345 {
			t.Errorf("父子级联依赖解析失败: %s, %d", s, i)
		}
	})

	if err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}

	// 验证父容器无法访问子容器的值
	_, err = parent.Invoke(func(i int) {})
	if err == nil {
		t.Errorf("父容器不应能访问子容器的数据")
	}
}

func TestInjector_Apply(t *testing.T) {
	inj := New()
	db := &mockDB{Name: "SQLite"}
	cfg := &AppConfig{AppName: "EmbeddedApp", Port: 9000}

	inj.MapTo(db, (*DatabaseService)(nil))
	inj.Map(cfg)

	ctrl := &Controller{Unused: "original"}
	err := inj.Apply(ctrl)
	if err != nil {
		t.Fatalf("Apply 失败: %v", err)
	}

	if ctrl.DB == nil || ctrl.DB.Query() != "result from SQLite" {
		t.Errorf("结构体 DB 注入失败")
	}
	if ctrl.Config == nil || ctrl.Config.AppName != "EmbeddedApp" {
		t.Errorf("结构体 Config 注入失败")
	}
	if ctrl.Unused != "original" {
		t.Errorf("非 inject 字段被意外篡改")
	}
}

func TestInjector_Errors(t *testing.T) {
	inj := New()

	// 传入非函数给 Invoke
	_, err := inj.Invoke("not a func")
	if err == nil {
		t.Errorf("传入非函数时应该返回错误")
	}

	// 缺少依赖参数
	_, err = inj.Invoke(func(nonExistent int) {})
	if err == nil {
		t.Errorf("缺少依赖时应该返回错误")
	}

	// MapTo 非接口指针
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MapTo 非接口指针应当 panic")
		}
	}()
	inj.MapTo(&mockDB{}, "not_interface_ptr")
}
