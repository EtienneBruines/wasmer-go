package wasmer

// #include <wasmer_wasm.h>
import "C"
import (
	"fmt"
	"unsafe"
)

type ImportObject struct {
	externs       map[string]map[string]IntoExtern
	opaqueExterns []IntoExtern
}

func NewImportObject() *ImportObject {
	return &ImportObject{
		externs:       make(map[string]map[string]IntoExtern),
		opaqueExterns: nil,
	}
}

func (self *ImportObject) intoInner(module *Module) (*C.wasm_extern_vec_t, error) {
	cExterns := &C.wasm_extern_vec_t{}

	/*
		// TODO(ivan): review this
			runtime.SetFinalizer(cExterns, func(cExterns *C.wasm_extern_vec_t) {
				C.wasm_extern_vec_delete(cExterns)
			})
	*/

	var externs []*C.wasm_extern_t
	var numberOfExterns uint

	for _, importType := range module.Imports() {
		namespace := importType.Module()
		name := importType.Name()

		if self.externs[namespace][name] == nil {
			return nil, &Error{
				message: fmt.Sprintf("Missing import: `%s`.`%s`", namespace, name),
			}
		}

		externs = append(externs, self.externs[namespace][name].IntoExtern().inner())
		numberOfExterns++
	}

	for _, extern := range self.opaqueExterns {
		externs = append(externs, extern.IntoExtern().inner())
		numberOfExterns++
	}

	if numberOfExterns > 0 {
		C.wasm_extern_vec_new(cExterns, C.size_t(numberOfExterns), (**C.wasm_extern_t)(unsafe.Pointer(&externs[0])))
	}

	return cExterns, nil
}

func (self *ImportObject) ContainsNamespace(name string) bool {
	_, exists := self.externs[name]

	return exists
}

func (self *ImportObject) Register(namespaceName string, namespace map[string]IntoExtern) {
	_, exists := self.externs[namespaceName]

	if exists == false {
		self.externs[namespaceName] = namespace
	} else {
		for key, value := range namespace {
			self.externs[namespaceName][key] = value
		}
	}
}

func (self *ImportObject) addOpaqueExtern(extern *Extern) {
	self.opaqueExterns = append(self.opaqueExterns, extern)
}
