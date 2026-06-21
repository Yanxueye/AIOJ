# TerminalOJ 前端重新设计提示词

## 当前问题分析

### 1. 样例输入框样式问题
- **当前**：黑色背景 (#1e1e1e)，与 LeetCode 不一致
- **LeetCode**：浅色背景 (#f5f5f5 或白色)，带边框，支持上下拖拽调整大小
- **修复**：改为浅色背景，添加拖拽调整功能

### 2. 整体视觉风格
- **当前**：暗色主题为主，但部分组件颜色不协调
- **目标**：统一的浅色主题，清新、专业的 OJ 风格

### 3. 布局问题
- 测试用例面板不能调整大小
- 编辑器和面板之间的分割线不够明显

## 设计提示词

```
重新设计 TerminalOJ 的题目详情页面，参考 LeetCode 的设计风格，但保持清新特色：

1. **测试用例面板**：
   - 背景色改为浅色 (#fafafa 或白色)
   - 输入框使用等宽字体，浅灰背景 (#f8f8f8)
   - 添加上下拖拽调整大小的功能（参考 LeetCode 的分割线）
   - Tab 栏使用简洁的下划线样式
   - 运行按钮使用绿色主题

2. **提交结果面板**：
   - 通过时显示绿色背景的 "Accepted" 横幅
   - 失败时显示红色背景，展示错误类型和失败用例
   - 复杂度信息使用卡片式展示
   - AI 分析按钮使用渐变色

3. **代码编辑器**：
   - 保持暗色主题（与 LeetCode 一致）
   - 工具栏使用深色背景
   - 语言选择器和字体大小控件

4. **整体配色**：
   - 主色调：绿色系 (#00b8a3 或 #2cbb5d)
   - 背景：浅灰 (#f5f5f5)
   - 卡片：白色
   - 文字：深灰 (#333)
   - 强调色：蓝色 (#1a73e8)

5. **交互细节**：
   - 测试用例面板支持折叠/展开
   - 拖拽分割线调整左右面板比例
   - 提交按钮使用绿色，运行按钮使用蓝色
   - 状态标签使用圆角样式
```

## 具体修改建议

### 测试用例面板样式修改

```css
/* 改为浅色主题 */
.tc-textarea {
  background: #f8f8f8;  /* 浅灰背景 */
  color: #333;          /* 深色文字 */
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  padding: 10px 12px;
  font-size: 13px;
  font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
  resize: vertical;
  outline: none;
  line-height: 1.6;
  width: 100%;
}

.testcase-body {
  background: #fafafa;  /* 浅色背景 */
  padding: 10px 14px 14px;
}

.testcase-header {
  background: #fff;     /* 白色背景 */
  border-top: 1px solid #e8e8e8;
}
```

### 添加拖拽调整功能

```vue
<!-- 在编辑器和测试用例面板之间添加可拖拽的分割线 -->
<div class="panel-divider" @mousedown="startResizePanel">
  <div class="divider-handle" />
</div>
```

```css
.panel-divider {
  height: 4px;
  cursor: row-resize;
  background: #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s;
}
.panel-divider:hover {
  background: #1a73e8;
}
.divider-handle {
  width: 30px;
  height: 3px;
  border-radius: 2px;
  background: #ccc;
}
```

### 状态标签样式

```css
/* 使用圆角标签样式 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}
.status-accepted { background: #dcfce7; color: #166534; }
.status-wrong { background: #fef2f2; color: #991b1b; }
.status-tle { background: #fff7ed; color: #9a3412; }
.status-ce { background: #fdf4ff; color: #86198f; }
```

## 参考资源

- LeetCode 题目详情页：https://leetcode.cn/problems/two-sum/
- Codeforces 提交页面：https://codeforces.com/submit
- VS Code 编辑器主题
