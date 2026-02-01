## 前期准备
- 需要安装配置好`cangjie-mem`

## 初始化仓颉基础语法书
- 下载[cj_syntax.md](https://github.com/ystyle/cangjie-docs-mcp/blob/master/cj_syntax.md)
- 让AI读取整个文件，并生成仓颉语言级记忆, 提示词:
  ```md
  读取cj_syntax.md分析并整理成按不同分类的记忆，然后记录到canmgjie-mem语言级记忆里
  ```

这样就生成了一份仓颉基础语法书的记忆， 在新项目或新会话中，使用提示词让ai加载就行了

`（这一步只做一次就行，一个cangjie-mem环境是可持久且多项目共用的，http版本还可以跨机器团队使用）`

## 使用-仓颉基础语法-记忆

在新项目或会话的使用示例, 提示词：
```md
使用cangjie-mem list加载所有的仓颉语言级记忆
```
或者同时指定需求`（需要提前写好设计文档）`
```md
 使用cangjie-mem list加载所有的仓颉语言级记忆, 并读取设计文档，和任务列表，然后按任务生成实现代码
```

**只要你觉得AI忘记了，也可以让它重新加载一下**

## 初始化三方库级记忆

>作用是，可以直接让ai加载某个库的记忆，而不需要每次使用某个库时让ai重新分析一下源码来让ai理解库的api使用法， 可以直接使用分析好的结果


1.  以`claude code`生成`tang`的api使用示例:
- 下载`git clone https://github.com/ystyle/tang`代码仓库
- 在`tang`项目启动`claude code`然后把以下提示词发给AI:
  ```md
  分析README.md、源码和docs下的文档，按不同的分类组织tang的资料并整理成cangje-mem的库级记忆，库名使用tang， 并全部记录到cangjie-mem里
  ```


2. 在使用`tang`的项目里，在新会话和ai说:
   ```md
   使用cangjie-mem list加载库名为tang的全部库级记， 然后写个hello world接口
   ```


## 推荐项目级提示词
把以下内容放`CLAUDE.md`里
```md
# [这里写项目名称]
>[这里写项目说明]

## 仓颉语言
- 在上下文压缩后，如果没有仓颉语法相关的，需要马上使用`cangjie_mem_list`工具加载所有仓颉语言级记忆
- 在仓颉api和手册可以使用`cangjie_docs`相关工具查找，在`cangjie-mem`没有的直接在文档里查找，不要猜api和语法
- 在提示语法错误时重新使用 `cangjie-mem` 加载语言级记忆
- match 的 case 后不能接`{}`, case后直接写多行列表式而不需要`{}`

## 任务指南
- **不要考虑时间，不要简化算法，不要简化测试，这项目是自己的产品，按最好的来搞**
- 需要在完成任务后更新`task.md`
- 在实现功能总结后，把总结的内容记录到`cangjie-mem`项目级记忆里
- 新功能、新特性一定要写单元测试
- 不要想在项目外创建仓颉单文件测试，非cjpm项目没法导入当前项目
- 新功能需要做好，且有单元测试后提交, 以仓颉单元测试为主
- 测试发现的新问题需要解决，且要添加新用例到仓颉的单元测试里

## 常用命令
\```shell
# 编译
cjpm build 
# 运行可执行包
cjpm run --name ystyle::xisp.examples.match_demo
# 手动执行，先查看 target/release/bin 目录有哪些可执行文件
./target/release/bin/ystyle::xisp.cli

# 运行所有测试
cjpm test 

# 运行指定测试
cjpm test --show-all-output --filter 'ParserTest.testParseRestParameter'
cjpm test --show-all-output --filter 'ParserTest.*'

# 清理构建缓存 
cjpm clean
\```
```
