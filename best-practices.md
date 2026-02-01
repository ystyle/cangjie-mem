### 前期准备
- 需要安装配置好`cangjie-mem`

### 初始化仓颉基础语法书
- 下载[cj_syntax.md](https://github.com/ystyle/cangjie-docs-mcp/blob/master/cj_syntax.md)
- 让AI读取整个文件，并生成仓颉语言级记忆, 提示词: `读取cj_syntax.md分析并整理成按不同分类的记忆，然后记录到canmgjie-mem语言级记忆里`

这样就生成了一份仓颉基础语法书的记忆， 在新项目或新会话中，使用提示词让ai加载就行了， 提示词：`使用cangjie-mem list加载所有的仓颉语言级记忆` 可以在这后面接`并读取设计文档，和任务表，然后按任务生成实现代码`（需要先写好设计文档）

### 初始化三方库级记忆
>作用是，可以直接让ai加载某个库的记忆，而不需要每次使用某个库时让ai重新分析一下源码来让ai理解库的api使用法， 可以直接使用分析好的结果


1.  以`claude code`生成`tang`的api使用示例:
- 下载`git clone https://github.com/ystyle/tang`代码仓库
- 在`tang`项目启动`claude code`然后把以下提示词发给AI:  `分析README.md、源码和docs下的文档，按不同的分类组织tang的资料并整理成cangje-mem的库级记忆，库名使用tang， 并全部记录到cangjie-mem里`


2. 在使用`tang`的项目里，在新会话和ai说: `使用cangjie-mem list加载库名为tang的全部库级记， 然后写个hello world接口`
