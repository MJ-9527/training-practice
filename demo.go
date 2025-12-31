package main

import (
	"encoding/json"
	"fmt"
)

type Animate struct {
	Name string
}

func (a Animate) run() {
	fmt.Printf("name:%v run\n", a.Name)

}

type Dog struct {
	Age int
	Animate
}

func (d Dog) wang() {
	fmt.Printf("name:%v wang wang\n", d.Name)
}

type User struct {
	Username string
	Password string
	Age      int
	// 匿名字段，类型Address
	Address
}

type Address struct {
	Name  string
	Phone string
	City  string
}

type Person struct {
	name string
	age  int
	sex  string
}

// 接收者类型可以是指针类型
func (p Person) PrintInfo() {
	fmt.Printf("姓名：%v, 年龄：%v\n", p.name, p.age)
}

func (p *Person) SetInfo(name string, age int) {
	p.name = name
	p.age = age
}

func main() {
	/*
		s := "你好，我是马骏"
		fmt.Println("什么意思！")

		for _, v := range s {
			fmt.Printf("(%c)", v)
		}

		rune1 := []rune(s)
		rune1[0] = '不'
		fmt.Println(string(rune1))

		var i int = 10
		// str := fmt.Sprintf("%d", i)
		// fmt.Printf("%v, %T", str, str) //10 string

		str1 := strconv.FormatInt(int64(i), 10)
		fmt.Printf("%v,%T", str1, str1)

		// 1. 数组的初始化
		var arr2 [3]int
		arr2[0] = 1
		arr2[1] = 2
		arr2[2] = 3
		fmt.Println(arr2) // [1 2 3]

		// 2. 数组的初始化
		var arr3 = [3]int{1, 2, 3}

		fmt.Println(arr3) // [1 2 3]

		// 3. 数组的初始化, 自动推断数组的长度
		var arr4 = [...]int{1, 2, 3, 4, 5}

		fmt.Printf("值：%v, 长度:%v", arr4, len(arr4)) // 值：[1 2 3 4 5], 长度:5

		// 4. 数组的初始化,索引的方式
		arr5 := [...]int{0: 1, 1: 10, 2: 20}
		fmt.Println(arr5) // [1 10 20]

		var arr6 = [...]int{0: 1, 1: 100, 2: 200}
		fmt.Println(arr6)

		for _, v := range arr6 {
			fmt.Println(v)
		}

		var arr7 = [3][2]string{{"北京", "上海"}, {"广州", "深圳"}, {"成都", "重庆"}}
		for _, v := range arr7 {
			fmt.Println(v)
		}

		// 切片
		var arr []int
		// arr[0] = 1  切片只是声明，长度为0，不但能通过这样赋值，赋值需要用append
		fmt.Printf("%v , %T, %v\n", arr, arr, len(arr)) //[] , []int, 0
		fmt.Println(arr == nil)                         // true, 声明切片没有赋值，默认切片的默认值为nil

		var q = []int{1, 2, 3}

		fmt.Printf("%v , %T, %v\n", q, q, len(q)) //[1 2 3] , []int, 3

		q[0] = 100 // 有切片长度，在切片长度内是可以通过这样赋值的
		fmt.Println(q)

		// 基于数组定义切片

		var s0 = [5]int{1, 2, 3, 4, 5}

		// 代表起始索引和结束索引的下一个位置但不包括这个位置
		s1 := s0[2:5]                                //
		fmt.Printf("%v , %T, %v\n", s1, s1, len(s1)) // [3 4 5] , []int, 3

		// 省略结束索引
		s2 := s0[1:]

		// 省略开始索引
		s3 := s0[:3]

		// 省略所有索引，复制原数组
		s4 := s0[:]

		fmt.Println(s2, s3, s4) // [2 3 4 5] [1 2 3] [1 2 3 4 5]

		// 切片是数组的"视图"，不存储数据，修改切片相当于修改数组
		s4[0] = 101
		fmt.Println(s0) //[100 2 3 4 5]

		// 切片的容量就是从它的第一个元素开始数，到其底层数组元素末尾的个数
		var ws = []int{1, 2, 3, 4, 5, 6}
		w2 := ws[2:]
		fmt.Printf("len:%v, cap:%v\n", len(w2), cap(w2)) // len:4 cap:4

		w3 := ws[1:3]
		fmt.Printf("len:%v, cap:%v\n", len(w3), cap(w3)) // len:2 cap:5

		w4 := ws[:3]
		fmt.Printf("len:%v, cap:%v\n", len(w4), cap(w4)) // len:3 cap:6

		fmt.Println(w2, w3, w4) //[3 4 5 6] [2 3] [1 2 3]

		// 使用make()构造切片
		var slice = make([]int, 4)

		fmt.Println(slice)                                    //[0 0 0 0]
		fmt.Printf("len:%v, cap: %v", len(slice), cap(slice)) //len:4, cap: 4

		// 合并切片，使用append();
		slice = append(slice, 10)
		fmt.Printf("%v, len:%v, cap: %v", slice, len(slice), cap(slice)) //[0 0 0 0 10], len:5, cap: 8
		// append 需要存放第 5 个元素，超过了当前容量，
		// Go 运行时会为切片分配新的底层数组并拷贝数据。
		// Go 对容量的增长有策略：对于小容量通常按 2 倍扩展（4 → 8），对于较大容量采用约 1.25 倍的增长

		// 这个同样也是运用这个原理
		var sliceC []int
		for i := 1; i < 10; i++ {
			sliceC = append(sliceC, i)
			fmt.Printf("%v, len: %v, cap: %v\n", sliceC, len(sliceC), cap(sliceC))
		}

		sliceA := []string{"php", "C"}
		sliceB := []string{"golang", "java"}

		sliceA = append(sliceA, sliceB...) // 切片合并需要加...
		fmt.Println(sliceA)                //[php C golang java]

		// // 输入使用
		// var name string
		// fmt.Println("请输入内容：")
		// fmt.Scan(&name) //传变量地址
		// fmt.Println("输入的内容是：", name)

		// 切片拷贝
		sliceD := []string{"php", "C"}
		sliceE := make([]string, 2)

		copy(sliceE, sliceD)
		fmt.Printf("%v, %v\n", sliceD, sliceE) // [php C], [php C]

		// 修改切片不会影响原切片
		sliceE[0] = "C++"
		fmt.Printf("%v, %v", sliceD, sliceE) //[php C], [C++ C]

		// map的使用
		var user = make(map[string]string)
		user["username"] = "zhangshan"
		user["age"] = "20"

		fmt.Println(user) // map[age:20 username:zhangshan]

		userinfo := map[string]string{
			"username": "zhansan",
			"age":      "20",
		}

		fmt.Println(userinfo)
		for k, v := range userinfo {
			fmt.Printf("key:%v, value:%v\n", k, v)
		}

		// 判断是否age存在
		v, ok := userinfo["age"]
		fmt.Println(v, ok)

		//基于map定义切片
		var userinfo2 = make([]map[string]string, 3)

		var person1 = make(map[string]string)
		person1["username"] = "zs"
		person1["age"] = "10"

		var person2 = make(map[string]string)
		person2["username"] = "ls"
		person2["age"] = "20"

		userinfo2[0] = person1
		userinfo2[1] = person2

		fmt.Println(userinfo2) // [map[age:10 username:zs] map[age:20 username:ls] map[]]

		// 循环切片中的map
		for _, v := range userinfo2 {
			for key, value := range v {
				fmt.Printf("key:%v, value: %v \n", key, value)
			}
		}


		var userinfo3 = make(map[string][]string)

		userinfo3["hobby"] = []string{
			"吃饭",
			"睡觉",
		}
		fmt.Println(userinfo3) // map[hobby:[吃饭 睡觉]]

		// map通过切片排序输出
		var map1 = make(map[int]int)

		map1[10] = 100
		map1[5] = 50
		map1[1] = 10
		map1[3] = 30
		map1[7] = 70
		map1[2] = 20

		fmt.Println(map1) //输出有序，实际迭代是无序的 map[1:10 2:20 3:30 5:50 7:70 10:100]

		// 1. 放入map key的值到一个切片中
		var slicenew []int
		for k, _ := range map1 {
			slicenew = append(slicenew, k)
		}

		fmt.Println(slicenew)

		// 2. 在切片中算出升序,选择排序或者冒泡排序
		for i := 0; i < len(slicenew); i++ {
			for j := i + 1; j < len(slicenew); j++ {
				if slicenew[i] > slicenew[j] {
					slicenew[i], slicenew[j] = slicenew[j], slicenew[i]
				}
			}
		}
		fmt.Println(slicenew)
		// 3. 输出map的值

		for _, v := range slicenew {
			fmt.Printf("map key:%v, value:%v\n", v, map1[v])
		}

		// 创建自定义类型
		type myInt int

		var a myInt = 10
		fmt.Printf("a: %v , a type:%T\n", a, a) // a: 10 , a type:main.myInt
		// 与原来的int类型不一样，需要转换才能使用
		var b int = int(a)
		fmt.Printf("b: %v , b type:%T\n", b, b) // b: 10 , b type:int

		// 起别名
		type myFloat = float64

		var a1 myFloat = 10
		fmt.Printf("a1: %v , a1 type:%T\n", a1, a1) // a1: 10 , a1 type:float64
		// 本质还是 float64 类型
		var b1 float64 = a1
		fmt.Printf("b1: %v , b1 type:%T\n", b1, b1) // b1: 10 , b1 type:float64
	*/

	//结构体的多种实例化方式
	var p1 Person // 实例化Person结构体，返回结构体本身
	p1.name = "zs"
	p1.sex = "男"
	p1.age = 20
	fmt.Printf("p1: %v , type:%T\n", p1, p1)
	fmt.Printf("p1: %#v, type:%T\n", p1, p1)
	/*
		p1: {zs 20 男} , type:main.Person
		p1: main.Person{name:"zs", age:20, sex:"男"}, type:main.Person
	*/

	var p2 = new(Person) // new()实例化结构体，返回的是结构体的地址，也就是指针类型
	p2.name = "ls"
	p2.sex = "女"
	p2.age = 18
	fmt.Printf("p2: %v , type:%T\n", p2, p2)
	fmt.Printf("p2: %#v, type:%T\n", p2, p2)
	/*
		p2: &{ls 18 女} , type:*main.Person
		p2: &main.Person{name:"ls", age:18, sex:"女"}, type:*main.Person
	*/

	var p3 = &Person{} // 取地址符&实例化结构体，返回结构体的地址，也就是指针类型
	p3.name = "lxxs"
	p3.sex = "女"
	p3.age = 11
	fmt.Printf("p3: %v , type:%T\n", p3, p3)
	fmt.Printf("p3: %#v, type:%T\n", p3, p3)
	/*
		p3: &{lxxs 11 女} , type:*main.Person
		p3: &main.Person{name:"lxxs", age:11, sex:"女"}, type:*main.Person
	*/

	var p4 = Person{
		name: "x嘻嘻嘻",
		age:  10,
		sex:  "男",
	} // 结构体字面量初始化，返回结构体本身
	fmt.Printf("p4: %v , type:%T\n", p4, p4)
	fmt.Printf("p4: %#v, type:%T\n", p4, p4)
	/*
		p4: {x嘻嘻嘻 10 男} , type:main.Person
		p4: main.Person{name:"x嘻嘻嘻", age:10, sex:"男"}, type:main.Person
	*/

	p5 := p4 // 结构体赋值，拷贝一份新的结构体给p5
	p5.name = "aaaa"
	fmt.Printf("p4: %v\n", p4) // p4: {x嘻嘻嘻 10 男}
	fmt.Printf("p5: %v\n", p5) // p5: {aaaa 10 男}

	// 给结构体添加方法
	// 方法是作用在特定类型上的函数
	// func (接收者变量 接收者类型) 方法名(参数列表) (返回参数) {方法体}

	p5.PrintInfo() // 姓名：aaaa, 年龄：10

	p5.SetInfo("xxx", 28)
	p5.PrintInfo() // 姓名：xxx, 年龄：28

	//结构体的嵌套
	var user User
	user.Username = "zs"
	user.Password = "passwoddo"
	user.Age = 18
	user.Address.Name = "caoyang"
	user.Address.City = "beijin"
	user.Address.Phone = "1888888"

	fmt.Printf("%#v\n", user)

	//结构体的继承
	var d = Dog{
		Age: 5,
		Animate: Animate{
			Name: "小黑",
		},
	}
	d.wang() // name:小黑 wang wang
	d.run()  // name:小黑, run

	//序列化和反序列化
	userinfo := User{
		Username: "zhangsan",
		Password: "password123",
		Age:      30,
		Address: Address{
			Name:  "home",
			Phone: "1234567890",
			City:  "Beijing",
		},
	}
	data, err := json.Marshal(userinfo)
	if err != nil {
		fmt.Println("序列化错误:", err)
		return
	}
	fmt.Println("序列化结果:", string(data))
	var userinfo2 User
	err = json.Unmarshal(data, &userinfo2)
	if err != nil {
		fmt.Println("反序列化错误:", err)
		return
	}
	fmt.Printf("反序列化结果: %#v\n", userinfo2)

	//结构体的Tag标签
	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	var product = Product{
		ID:    1,
		Name:  "Laptop",
		Price: 999.99,
	}
	productData, err := json.Marshal(product)
	if err != nil {
		fmt.Println("序列化错误:", err)
		return
	}
	fmt.Println("带Tag标签的序列化结果:", string(productData))

}
