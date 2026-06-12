# TerminalOJ 前端风格指南

## 设计方向：清新二次元 + 极简代码美学

### 核心理念
将算法竞赛的严谨与二次元的活力结合，打造一个让程序员感到舒适、有动力刷题的界面。

### 配色方案

```css
/* 主色调 - 薄荷绿 */
--primary: #00b894;
--primary-light: #55efc4;
--primary-dark: #00a882;

/* 强调色 - 珊瑚橙 */
--accent: #ff6b6b;
--accent-light: #ffa8a8;

/* 背景 - 温暖米白 */
--bg-main: #faf9f6;
--bg-card: #ffffff;
--bg-hover: #f5f3ee;

/* 文字 */
--text-primary: #2d3436;
--text-secondary: #636e72;
--text-muted: #b2bec3;

/* 暗色模式 */
--dark-bg: #1a1a2e;
--dark-card: #16213e;
--dark-accent: #0f3460;
```

### 标语方案

**当前（需要替换）：**
```
刷题，不止于刷题
```

**建议方案：**
```
代码即诗意，算法即远方
```
或
```
每一行代码，都是向 AC 迈进的一步
```
或
```
Debug 人生，Compile 未来
```

### 背景图片适配

使用 CSS 实现二次元风格背景：
- 浅色模式：淡淡的渐变网格 + 模糊的代码片段装饰
- 暗色模式：深色星空渐变 + 霓虹色调的代码高亮

### 卡片设计

- 圆角：16px（大卡片）、12px（小卡片）
- 阴影：柔和的多层阴影
- 悬停：轻微上浮 + 阴影加深
- 边框：1px solid rgba(0,0,0,0.06)

### 字体

- 标题：'Noto Sans SC', 'PingFang SC', sans-serif（中文）
- 正文：'Inter', 'SF Pro Text', sans-serif
- 代码：'JetBrains Mono', 'Fira Code', monospace
