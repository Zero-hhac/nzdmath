import fs from 'fs';
import path from 'path';

const BASE_URL = 'http://localhost:8080/api/v1/admin';

async function seedLatex() {
  console.log('Logging in as admin...');
  const loginRes = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' })
  });
  if (!loginRes.ok) throw new Error('Login failed');
  const loginData = await loginRes.json();
  const token = loginData.data.token;

  console.log('Inserting rich LaTeX article...');
  const res = await fetch(`${BASE_URL}/news`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      title: 'LaTeX排版入门：从零写出优雅的数学论文（终极修订版）',
      summary: '本文面向LaTeX零基础读者，从安装配置开始，逐步讲解数学公式、定理环境、参考文献管理等核心功能。包含丰富示例代码。',
      category: '学习资源',
      tag: 'LaTeX, 排版, 教程',
      cover_url: '',
      is_featured: true,
      status: 1,
      content: `
# LaTeX 排版入门：从零写出优雅的数学论文

LaTeX 是一种基于 TeX 的排版系统。它广泛用于生成高质量的科学和数学文档。与 Word 不同，LaTeX 让你专注于**内容**本身，而将排版工作交给系统。

## 1. 为什么选择 LaTeX？
*   **无与伦比的数学公式排版**
*   **自动化编号**（章节、公式、图表）
*   **完美的参考文献管理**
*   **跨平台且开源免费**

## 2. 核心语法基础

在 LaTeX 中，你的文档是由纯文本和命令组成的。比如，创建一个最基本的文档骨架：

\`\`\`latex
\\documentclass{article}
\\usepackage{amsmath}

\\begin{document}
Hello, world! 这是一个简单的 LaTeX 示例文档。
\\end{document}
\`\`\`

## 3. 数学公式的艺术

LaTeX 最为人称道的就是其对复杂数学表达式的渲染。我们可以使用单美元符号 \`$...\$\` 来插入行内公式，或者使用双美元符号或 \`\\[ ... \\]\` 来插入行间公式。

### 3.1 行内公式
伟大的欧拉公式 $e^{i\\pi} + 1 = 0$ 是数学中最美的等式之一。它将数学中最重要的五个常数 $0, 1, e, i, \\pi$ 巧妙地联系在了一起。

### 3.2 行间公式
我们可以轻松排版复杂的积分或级数。例如，黎曼 Zeta 函数的定义：

$$ \\zeta(s) = \\sum_{n=1}^{\\infty} \\frac{1}{n^s} $$

或者是高斯积分：

$$ \\int_{-\\infty}^{\\infty} e^{-x^2} \\, dx = \\sqrt{\\pi} $$

> 注意：渲染上面的公式需要我们在前端配置好 \`remark-math\` 和 \`rehype-katex\`！

### 3.3 矩阵排版
在 \`amsmath\` 宏包下，排版矩阵也非常简单：

$$
A = \\begin{pmatrix}
a_{11} & a_{12} & a_{13} \\\\
a_{21} & a_{22} & a_{23} \\\\
a_{31} & a_{32} & a_{33}
\\end{pmatrix}
$$

## 4. 定理与证明环境

在数学论文中，我们经常需要书写定理和证明。我们可以使用 \`amsthm\` 宏包：

\`\`\`latex
\\newtheorem{theorem}{定理}
\\begin{theorem}[费马大定理]
当整数 $n > 2$ 时，关于 $x, y, z$ 的方程 $x^n + y^n = z^n$ 没有正整数解。
\\end{theorem}
\\begin{proof}
这个证明太长了，空白处写不下。
\\end{proof}
\`\`\`

## 5. 结语

LaTeX 的学习曲线可能一开始比较陡峭，但只要掌握了基础，你就会发现它是无可替代的利器。希望这篇快速入门指南能帮你顺利开启优雅的排版之旅！
`
    })
  });
  if (res.ok) {
    console.log('Successfully inserted rich LaTeX article.');
  } else {
    console.error('Failed to insert article:', await res.text());
  }
}

seedLatex().catch(console.error);
