package reflect

import "reflect"

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	typ := reflect.TypeOf(entity)
	numMethod := typ.NumMethod()
	res := make(map[string]FuncInfo, numMethod)

	for i := 0; i < numMethod; i++ {
		method := typ.Method(i)
		fn := method.Func

		numIn := fn.Type().NumIn()
		input := make([]reflect.Type, 0, numIn)
		inputValues := make([]reflect.Value, 0, numIn)
		for j := 0; j < numIn; j++ {
			fnInType := fn.Type().In(j)
			input = append(input, fnInType)
			inputValues = append(inputValues, reflect.Zero(fnInType))
		}

		numOut := fn.Type().NumOut()
		output := make([]reflect.Type, 0, numOut)
		for i := 0; i < numOut; i++ {
			output = append(output, fn.Type().Out(i))
		}

		resValues := fn.Call(inputValues)
		result := make([]any, 0, len(resValues))
		for _, val := range resValues {
			result = append(result, val.Interface())
		}
		res[method.Name] = FuncInfo{
			Name:        method.Name,
			InputTypes:  input,
			OutputTypes: output,
			Result:      result,
		}
	}
	return res, nil
}

type FuncInfo struct {
	Name        string
	InputTypes  []reflect.Type
	OutputTypes []reflect.Type
	Result      []any
}
