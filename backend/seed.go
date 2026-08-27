package main

import (
	"encoding/json"

	"codelearn/model"
	"codelearn/repository"
)

func seedData(repo *repository.Repository) error {
	var count int64
	repo.DB().Model(&model.Course{}).Count(&count)
	if count > 0 {
		return nil
	}

	courses := []model.Course{
		{Language: "go", Title: "Go 语言入门", Description: "从零开始学习 Go 语言编程", Emoji: "🐹", Color: "#00ADD8", Order: 1},
		{Language: "python", Title: "Python 编程入门", Description: "掌握 Python 编程基础", Emoji: "🐍", Color: "#3776AB", Order: 2},
	}
	if err := repo.DB().Create(&courses).Error; err != nil {
		return err
	}

	// === Go 课程 ===
	goUnits := []model.Unit{
		{CourseID: courses[0].ID, Title: "基础语法", Description: "Go 语言的基本语法", Icon: "📚", Color: "#00ADD8", Order: 1},
		{CourseID: courses[0].ID, Title: "控制流与函数", Description: "条件判断、循环和函数", Icon: "🔀", Color: "#00ADD8", Order: 2},
		{CourseID: courses[0].ID, Title: "数据结构", Description: "数组、切片、映射和结构体", Icon: "🗂️", Color: "#00ADD8", Order: 3},
	}
	if err := repo.DB().Create(&goUnits).Error; err != nil {
		return err
	}

	goLessons := []model.Lesson{
		{UnitID: goUnits[0].ID, Title: "Hello World", Description: "Go 程序的基本结构", Content: "# Go 程序结构\n\n每个 Go 程序都从 package main 开始，main 函数是程序入口。\n\n示例代码:\n    package main\n    import \"fmt\"\n    func main() {\n        fmt.Println(\"Hello, World!\")\n    }\n\n关键概念:\n- package main 声明这是一个可执行程序\n- import \"fmt\" 导入格式化输出包\n- func main() 是程序入口", Icon: "👋", Order: 1},
		{UnitID: goUnits[0].ID, Title: "变量与类型", Description: "声明变量和基本数据类型", Content: "# 变量与类型\n\nGo 是静态类型语言，变量声明方式:\n    var name string = \"Go\"\n    age := 18  // 短变量声明\n    const Pi = 3.14159\n\n基本类型: int, float64, string, bool", Icon: "📦", Order: 2},
		{UnitID: goUnits[1].ID, Title: "条件判断", Description: "if-else 和 switch", Content: "# 条件判断\n\nGo 的 if 语句不需要括号:\n    if age >= 18 {\n        fmt.Println(\"成年人\")\n    } else {\n        fmt.Println(\"未成年\")\n    }", Icon: "⚖️", Order: 1},
		{UnitID: goUnits[1].ID, Title: "循环", Description: "for 循环", Content: "# 循环\n\nGo 只有 for 一种循环:\n    for i := 0; i < 5; i++ {\n        fmt.Println(i)\n    }", Icon: "🔄", Order: 2},
		{UnitID: goUnits[1].ID, Title: "函数", Description: "定义和调用函数", Content: "# 函数\n\nGo 函数定义示例:\n    func add(a, b int) int {\n        return a + b\n    }\n\nGo 支持多返回值:\n    func divmod(a, b int) (int, int) {\n        return a / b, a % b\n    }", Icon: "🔧", Order: 3},
		{UnitID: goUnits[2].ID, Title: "切片", Description: "动态数组", Content: "# 切片\n\n切片是 Go 中最常用的数据结构:\n    nums := []int{1, 2, 3}\n    nums = append(nums, 4)\n    fmt.Println(len(nums)) // 4", Icon: "✂️", Order: 1},
		{UnitID: goUnits[2].ID, Title: "映射 Map", Description: "键值对", Content: "# Map\n\nMap 是键值对集合:\n    m := map[string]int{\"a\": 1, \"b\": 2}\n    m[\"c\"] = 3\n    fmt.Println(m[\"a\"]) // 1", Icon: "🗺️", Order: 2},
		{UnitID: goUnits[2].ID, Title: "结构体", Description: "自定义类型", Content: "# 结构体\n\n结构体是 Go 的自定义类型:\n    type Person struct {\n        Name string\n        Age  int\n    }\n    p := Person{Name: \"Alice\", Age: 25}", Icon: "🏗️", Order: 3},
	}
	if err := repo.DB().Create(&goLessons).Error; err != nil {
		return err
	}

	goExercises := buildGoExercises(goLessons)
	if err := repo.DB().Create(&goExercises).Error; err != nil {
		return err
	}

	// === Python 课程 ===
	pyUnits := []model.Unit{
		{CourseID: courses[1].ID, Title: "基础语法", Description: "Python 基本语法", Icon: "📚", Color: "#3776AB", Order: 1},
		{CourseID: courses[1].ID, Title: "控制流与函数", Description: "条件、循环和函数", Icon: "🔀", Color: "#3776AB", Order: 2},
		{CourseID: courses[1].ID, Title: "数据结构", Description: "列表、字典、集合", Icon: "🗂️", Color: "#3776AB", Order: 3},
	}
	if err := repo.DB().Create(&pyUnits).Error; err != nil {
		return err
	}

	pyLessons := []model.Lesson{
		{UnitID: pyUnits[0].ID, Title: "Hello World", Description: "Python 程序入门", Content: "# Hello World\n\nPython 不需要 main 函数，直接执行脚本即可:\n    print(\"Hello, World!\")\n\nprint() 是 Python 的内置输出函数。", Icon: "👋", Order: 1},
		{UnitID: pyUnits[0].ID, Title: "变量与类型", Description: "动态类型语言", Content: "# 变量与类型\n\nPython 是动态类型语言:\n    name = \"Python\"\n    age = 18\n    pi = 3.14\n    is_active = True\n\n基本类型: int, float, str, bool", Icon: "📦", Order: 2},
		{UnitID: pyUnits[1].ID, Title: "条件判断", Description: "if-elif-else", Content: "# 条件判断\n\nPython 使用缩进表示代码块:\n    if age >= 18:\n        print(\"成年人\")\n    else:\n        print(\"未成年\")", Icon: "⚖️", Order: 1},
		{UnitID: pyUnits[1].ID, Title: "循环", Description: "for 和 while", Content: "# 循环\n\nPython 有 for 和 while 两种循环:\n    for i in range(5):\n        print(i)\n\nrange(5) 生成 0 到 4 共 5 个数字。", Icon: "🔄", Order: 2},
		{UnitID: pyUnits[1].ID, Title: "函数", Description: "def 定义函数", Content: "# 函数\n\n使用 def 关键字定义函数:\n    def add(a, b):\n        return a + b\n\nPython 函数支持默认参数和关键字参数。", Icon: "🔧", Order: 3},
		{UnitID: pyUnits[2].ID, Title: "列表", Description: "可变序列", Content: "# 列表\n\n列表是 Python 最常用的数据结构:\n    nums = [1, 2, 3]\n    nums.append(4)\n    print(nums[0]) # 1", Icon: "📋", Order: 1},
		{UnitID: pyUnits[2].ID, Title: "字典", Description: "键值对", Content: "# 字典\n\n字典是键值对集合:\n    d = {\"a\": 1, \"b\": 2}\n    d[\"c\"] = 3\n    print(d[\"a\"]) # 1", Icon: "📖", Order: 2},
		{UnitID: pyUnits[2].ID, Title: "集合", Description: "无序不重复", Content: "# 集合\n\n集合是无序且不重复的元素集合:\n    s = {1, 2, 3}\n    s.add(4)\n    s.add(1) # 不会重复添加", Icon: "🔢", Order: 3},
	}
	if err := repo.DB().Create(&pyLessons).Error; err != nil {
		return err
	}

	pyExercises := buildPythonExercises(pyLessons)
	if err := repo.DB().Create(&pyExercises).Error; err != nil {
		return err
	}

	return nil
}

func opts(s ...string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func tc1(in1, exp1 string) string {
	b, _ := json.Marshal([]map[string]string{{"input": in1, "expected": exp1}})
	return string(b)
}

func tc(in1, exp1, in2, exp2 string) string {
	b, _ := json.Marshal([]map[string]string{
		{"input": in1, "expected": exp1},
		{"input": in2, "expected": exp2},
	})
	return string(b)
}

func buildGoExercises(lessons []model.Lesson) []model.Exercise {
	return []model.Exercise{
		{LessonID: lessons[0].ID, Type: "choice", Question: "Go 程序的入口包名是什么？", Options: opts("main", "start", "app", "program"), Answer: "main", Explanation: "每个可执行 Go 程序都必须包含 package main 声明。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[0].ID, Type: "fillblank", Question: "输出文本到标准输出需要导入 ___ 包。", Answer: "fmt", Explanation: "fmt 包提供了格式化输入输出功能，Println 是其中最常用的函数。", Difficulty: "easy", Order: 2},
		{LessonID: lessons[1].ID, Type: "choice", Question: "以下哪个是 Go 的短变量声明语法？", Options: opts("name := value", "name = value", "var name := value", "let name = value"), Answer: "name := value", Explanation: ":= 是短变量声明操作符，只能在函数内部使用，会自动推断类型。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[1].ID, Type: "fillblank", Question: "声明常量使用 ___ 关键字。", Answer: "const", Explanation: "const 用于声明编译时常量，值不可修改。", Difficulty: "easy", Order: 2},
		{LessonID: lessons[2].ID, Type: "choice", Question: "Go 的 if 语句条件需要括号吗？", Options: opts("不需要", "必须加", "可选", "看情况"), Answer: "不需要", Explanation: "Go 的 if 不需要小括号包裹条件，但大括号是必须的。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[2].ID, Type: "code", Question: "读取一个整数，如果大于等于 60 输出 pass，否则输出 fail。", CodeTemplate: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    var score int\n    fmt.Scan(&score)\n    // 在这里写你的代码\n}", TestCases: tc("60", "pass", "50", "fail"), Difficulty: "medium", Order: 2},
		{LessonID: lessons[3].ID, Type: "choice", Question: "Go 有哪几种循环语句？", Options: opts("只有 for", "for 和 while", "for、while 和 do-while", "for 和 foreach"), Answer: "只有 for", Explanation: "Go 简化了循环语法，只有 for 一种循环语句，可实现 while 效果。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[3].ID, Type: "code", Question: "读取整数 n，输出 1 到 n 的和。", CodeTemplate: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    var n int\n    fmt.Scan(&n)\n    // 在这里写你的代码\n}", TestCases: tc("5", "15", "10", "55"), Difficulty: "medium", Order: 2},
		{LessonID: lessons[4].ID, Type: "choice", Question: "Go 函数可以返回多个值吗？", Options: opts("可以", "不可以", "需要特殊声明", "仅 Go 1.18+"), Answer: "可以", Explanation: "Go 原生支持多返回值，这是 Go 的特色之一。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[4].ID, Type: "code", Question: "读取两个整数，输出它们的和。", CodeTemplate: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    var a, b int\n    fmt.Scan(&a, &b)\n    // 在这里写你的代码\n}", TestCases: tc("1 2", "3", "10 20", "30"), Difficulty: "easy", Order: 2},
		{LessonID: lessons[5].ID, Type: "choice", Question: "向切片添加元素使用什么函数？", Options: opts("append", "add", "push", "insert"), Answer: "append", Explanation: "append 是内置函数，用于向切片追加元素。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[5].ID, Type: "fillblank", Question: "创建一个空字符串切片：s := make([]___, 0)", Answer: "string", Explanation: "[]string 声明字符串类型的切片。", Difficulty: "easy", Order: 2},
		{LessonID: lessons[6].ID, Type: "choice", Question: "Go 中 Map 的键可以是以下哪种类型？", Options: opts("所有可比较类型", "任意类型", "仅字符串", "仅整数"), Answer: "所有可比较类型", Explanation: "Map 的键必须是可比较的类型，如 string、int、bool 等。", Difficulty: "medium", Order: 1},
		{LessonID: lessons[6].ID, Type: "code", Question: "读取一个整数 n，输出 n*2。", CodeTemplate: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    var n int\n    fmt.Scan(&n)\n    // 在这里写你的代码\n}", TestCases: tc("5", "10", "7", "14"), Difficulty: "easy", Order: 2},
		{LessonID: lessons[7].ID, Type: "choice", Question: "Go 使用什么关键字定义结构体？", Options: opts("struct", "class", "object", "record"), Answer: "struct", Explanation: "Go 使用 type 和 struct 关键字定义结构体，没有 class 关键字。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[7].ID, Type: "fillblank", Question: "定义结构体：type Person ___ { Name string }", Answer: "struct", Explanation: "type Name struct { ... } 是定义结构体的语法。", Difficulty: "easy", Order: 2},
	}
}

func buildPythonExercises(lessons []model.Lesson) []model.Exercise {
	return []model.Exercise{
		{LessonID: lessons[0].ID, Type: "choice", Question: "Python 中输出文本使用哪个函数？", Options: opts("print", "echo", "console.log", "printf"), Answer: "print", Explanation: "print() 是 Python 的内置输出函数。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[0].ID, Type: "code", Question: "输出 Hello World", CodeTemplate: "# 在这里写你的代码", TestCases: tc1("", "Hello World"), Difficulty: "easy", Order: 2},
		{LessonID: lessons[1].ID, Type: "choice", Question: "Python 声明变量需要指定类型吗？", Options: opts("不需要", "必须指定", "可选", "仅函数参数"), Answer: "不需要", Explanation: "Python 是动态类型语言，变量类型由赋值自动推断。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[1].ID, Type: "fillblank", Question: "查看变量类型的函数是 ___()", Answer: "type", Explanation: "type() 函数返回对象的类型。", Difficulty: "easy", Order: 2},
		{LessonID: lessons[2].ID, Type: "choice", Question: "Python 中 else if 的写法是？", Options: opts("elif", "elseif", "else if", "elsif"), Answer: "elif", Explanation: "Python 使用 elif 表示 else if，这是 Python 的独特写法。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[2].ID, Type: "code", Question: "读取一个整数，如果大于等于 60 输出 pass，否则输出 fail。", CodeTemplate: "n = int(input())\n# 在这里写你的代码", TestCases: tc("60", "pass", "50", "fail"), Difficulty: "medium", Order: 2},
		{LessonID: lessons[3].ID, Type: "choice", Question: "range(5) 生成哪些数字？", Options: opts("0 1 2 3 4", "1 2 3 4 5", "0 1 2 3 4 5", "1 2 3 4"), Answer: "0 1 2 3 4", Explanation: "range(5) 生成 0 到 4 共 5 个数字，不包含 5。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[3].ID, Type: "code", Question: "读取整数 n，输出 1 到 n 的和。", CodeTemplate: "n = int(input())\n# 在这里写你的代码", TestCases: tc("5", "15", "10", "55"), Difficulty: "medium", Order: 2},
		{LessonID: lessons[4].ID, Type: "choice", Question: "Python 定义函数使用哪个关键字？", Options: opts("def", "func", "function", "fn"), Answer: "def", Explanation: "def 关键字用于定义函数。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[4].ID, Type: "code", Question: "读取两个整数（空格分隔），输出它们的和。", CodeTemplate: "a, b = map(int, input().split())\n# 在这里写你的代码", TestCases: tc("1 2", "3", "10 20", "30"), Difficulty: "easy", Order: 2},
		{LessonID: lessons[5].ID, Type: "choice", Question: "向列表添加元素使用什么方法？", Options: opts("append", "add", "push", "insert"), Answer: "append", Explanation: "list.append() 在列表末尾添加元素。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[5].ID, Type: "code", Question: "读取整数 n，输出 n 的阶乘。", CodeTemplate: "n = int(input())\n# 在这里写你的代码", TestCases: tc("5", "120", "3", "6"), Difficulty: "medium", Order: 2},
		{LessonID: lessons[6].ID, Type: "choice", Question: "Python 中字典的键可以是？", Options: opts("不可变类型", "任意类型", "仅字符串", "仅整数"), Answer: "不可变类型", Explanation: "字典的键必须是可哈希的（不可变）类型，如字符串、数字、元组。", Difficulty: "medium", Order: 1},
		{LessonID: lessons[6].ID, Type: "fillblank", Question: "创建空字典：d = ___", Answer: "{}", Explanation: "{} 或 dict() 都可以创建空字典。", Difficulty: "easy", Order: 2},
		{LessonID: lessons[7].ID, Type: "choice", Question: "集合的特点是什么？", Options: opts("无序不重复", "有序可重复", "无序可重复", "有序不重复"), Answer: "无序不重复", Explanation: "集合是无序且不重复的元素集合。", Difficulty: "easy", Order: 1},
		{LessonID: lessons[7].ID, Type: "code", Question: "读取整数 n，输出 n 的平方。", CodeTemplate: "n = int(input())\n# 在这里写你的代码", TestCases: tc("5", "25", "7", "49"), Difficulty: "easy", Order: 2},
	}
}
