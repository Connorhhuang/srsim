package eval

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/simimpact/srsim/pkg/gcs/ast"
	"github.com/simimpact/srsim/pkg/key"
)

func (e *Eval) initSysFuncs(env *Env) {
	// std funcs
	e.addSysFunc("rand", e.rand, env)
	e.addSysFunc("randnorm", e.randnorm, env)
	e.addSysFunc("print", e.print, env)
	e.addSysFunc("type", e.typeval, env)
	e.addSysFunc("register_skill_cb", e.registerSkillCB, env)
	e.addSysFunc("register_burst_cb", e.registerBurstCB, env)

	// char should be key.TargetID (e.g. dummy_character = 0)
}

func (e *Eval) addSysFunc(name string, f func(c *ast.CallExpr, env *Env) (Obj, error), env *Env) {
	var obj Obj = &bfuncval{Body: f}
	env.varMap[name] = &obj
}

func (e *Eval) print(c *ast.CallExpr, env *Env) (Obj, error) {
	//concat all args
	var sb strings.Builder
	for _, arg := range c.Args {
		val, err := e.evalExpr(arg, env)
		if err != nil {
			return nil, err
		}
		sb.WriteString(val.Inspect())
	}
	fmt.Println(sb.String())
	return &null{}, nil
}

func (e *Eval) rand(c *ast.CallExpr, env *Env) (Obj, error) {
	x := rand.Float64() // TODO: rand with a specific seed
	return &number{
		fval:    x,
		isFloat: true,
	}, nil
}

func (e *Eval) randnorm(c *ast.CallExpr, env *Env) (Obj, error) {
	x := rand.NormFloat64() // TODO: rand with a specific seed
	return &number{
		fval:    x,
		isFloat: true,
	}, nil
}

func (e *Eval) typeval(c *ast.CallExpr, env *Env) (Obj, error) {
	//type(var)
	if len(c.Args) != 1 {
		return nil, fmt.Errorf("invalid number of params for type, expected 1 got %v", len(c.Args))
	}

	t, err := e.evalExpr(c.Args[0], env)
	if err != nil {
		return nil, err
	}

	str := "unknown"
	switch t.Typ() {
	case typNull:
		str = "null"
	case typNum:
		str = "number"
	case typStr:
		str = "string"
	case typMap:
		str = "map"
	case typFun:
		fallthrough
	case typBif:
		str = t.Inspect()
	}

	return &strval{str}, nil
}

func (e *Eval) registerSkillCB(c *ast.CallExpr, env *Env) (Obj, error) {
	//register_skill_cb(char, func)
	if len(c.Args) != 2 {
		return nil, fmt.Errorf("invalid number of params for register_skill_cb, expected 2 got %v", len(c.Args))
	}

	//should eval to a function
	tarobj, err := e.evalExpr(c.Args[0], env)
	if err != nil {
		return nil, err
	}
	if tarobj.Typ() != typNum {
		return nil, fmt.Errorf("register_skill_cb argument char should evaluate to a number, got %v", tarobj.Inspect())
	}
	target := tarobj.(*number).ival

	//should eval to a function
	funcobj, err := e.evalExpr(c.Args[1], env)
	if err != nil {
		return nil, err
	}
	if funcobj.Typ() != typFun {
		return nil, fmt.Errorf("register_skill_cb argument func should evaluate to a function, got %v", funcobj.Inspect())
	}
	fn := funcobj.(*funcval)

	node := TargetNode{
		target: key.TargetID(target),
		env:    NewEnv(env),
		node:   fn.Body,
	}
	for i, v := range fn.Args {
		param, err := e.evalExpr(c.Args[i], env)
		if err != nil {
			return nil, err
		}
		node.env.varMap[v.Value] = &param
	}
	e.targetNode[key.TargetID(target)] = node
	return &null{}, nil
}

func (e *Eval) registerBurstCB(c *ast.CallExpr, env *Env) (Obj, error) {
	//register_burst_cb(char, func)
	if len(c.Args) != 2 {
		return nil, fmt.Errorf("invalid number of params for register_burst_cb, expected 2 got %v", len(c.Args))
	}

	//should eval to a function
	tarobj, err := e.evalExpr(c.Args[0], env)
	if err != nil {
		return nil, err
	}
	if tarobj.Typ() != typNum {
		return nil, fmt.Errorf("register_burst_cb argument char should evaluate to a number, got %v", tarobj.Inspect())
	}
	target := tarobj.(*number).ival

	//should eval to a function
	funcobj, err := e.evalExpr(c.Args[1], env)
	if err != nil {
		return nil, err
	}
	if funcobj.Typ() != typFun {
		return nil, fmt.Errorf("register_burst_cb argument func should evaluate to a function, got %v", funcobj.Inspect())
	}
	fn := funcobj.(*funcval)

	node := TargetNode{
		target: key.TargetID(target),
		env:    NewEnv(env),
		node:   fn.Body,
	}
	for i, v := range fn.Args {
		param, err := e.evalExpr(c.Args[i], env)
		if err != nil {
			return nil, err
		}
		node.env.varMap[v.Value] = &param
	}
	e.burstNodes = append(e.burstNodes, node)
	return &null{}, nil
}
