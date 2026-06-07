import fs from 'fs';
import path from 'path';

const BASE_URL = 'http://localhost:8080/api/v1/admin';

async function seed() {
  console.log('Logging in as admin...');
  const loginRes = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' }) // Assuming default is admin:admin123 or similar. Let's try 123456 if it fails.
  });
  
  if(!loginRes.ok) {
     const loginRes2 = await fetch(`${BASE_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: '123456' })
     });
     if (!loginRes2.ok) {
        console.error('Login failed'); return;
     }
     var { data: { token } } = await loginRes2.json();
  } else {
     var { data: { token } } = await loginRes.json();
  }
  
  console.log('Seeding fake markdown data...');
  const headers = { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` };

  // 1. Post News
  await fetch(`${BASE_URL}/news`, { method: 'POST', headers, body: JSON.stringify({ title: '关于黎曼猜想的非正式讨论与阅读清单', summary: '本周末，数爱会将举办一场针对黎曼猜想的交流茶话会，附录有供大家提前预习的参考书单。', category: '讨论', tag: '黎曼猜想', status: 1, is_featured: true, content: "# 黎曼猜想 (Riemann Hypothesis) 阅读清单\n\n> 黎曼猜想是数学中最重要的未解问题之一。它不仅涉及素数的分布规律，也深深影响着现代数论的脉络。\n\n以下是我们将要在讨论会上分享和探讨的核心议题：\n\n## 1. 黎曼 Zeta 函数\n\n黎曼 Zeta 函数在复平面上的定义为：\n对于实部大于 1 的复数 s，定义为 $\\zeta(s) = \\sum_{n=1}^{\\infty} \\frac{1}{n^s}$。\n\n## 2. 核心猜想陈述\n\n**黎曼猜想断言：** 黎曼 Zeta 函数的所有非平凡零点，其解析延拓后的实部均为 `1/2`。\n这意味着，所有非平凡零点都落在复平面的**临界线**上。\n\n### 推荐阅读书目\n- 《素数之恋》 (Prime Obsession) - John Derbyshire\n- 《黎曼猜想漫谈》 - 卢昌海\n\n### 讨论会安排\n* **时间**：本周五下午 3:00 - 5:00\n* **地点**：图书馆 3 层数学沙龙区\n* **要求**：请大家提前阅读上述书目的第一章。\n\n我们非常期待你的加入，一起探讨这个迷人而深邃的话题！" }) });

  // 2. Post Event
  await fetch(`${BASE_URL}/events`, { method: 'POST', headers, body: JSON.stringify({ title: '计算拓扑学前沿讲座：从理论到代码', summary: '我们将介绍单纯复形、持久同调等概念，并使用 Python 进行现场演示。', category: '讲座', location: '理科楼 A401', start_time: new Date(Date.now() + 86400000 * 3).toISOString(), end_time: new Date(Date.now() + 86400000 * 3 + 7200000).toISOString(), status: 1, is_featured: true, content: "# 计算拓扑学：持久同调入门\n\n在数据科学和机器学习中，提取高维数据的形状特征变得越来越重要。这就是**计算拓扑学**大显身手的地方。\n\n## 什么是持久同调 (Persistent Homology)？\n\n它是拓扑数据分析 (TDA) 中的核心工具，通过在不同的尺度下构建拓扑结构（如单纯复形），来观察哪些拓扑特征（孔洞、连通分支）是短暂的噪音，哪些是“持久”的本质结构。\n\n### 代码演示预告\n\n在讲座中，我们将使用 Python 的 `Giotto-TDA` 和 `Ripser` 库来进行演示：\n\n```python\nimport numpy as np\nfrom ripser import ripser\nfrom persim import plot_diagrams\n\n# 生成一个圆上的噪声数据\ndata = np.random.random((100, 2))\n\n# 计算持久同调\ndiagrams = ripser(data)['dgms']\n\n# 绘制条形图\nplot_diagrams(diagrams, show=True)\n```\n\n我们现场见！" }) });

  // 3. Post Showcase
  await fetch(`${BASE_URL}/showcases`, { method: 'POST', headers, body: JSON.stringify({ title: '分形几何的可视化应用', author: '刘洋', field: '几何学', competition: '国赛', summary: '这篇作品探讨了如何通过递归算法高效绘制曼德勃罗集，并在校内计算机应用大赛中获得一等奖。', status: 1, content: "# 分形几何的无限嵌套\n\n分形之美在于其自相似性。这段作品主要聚焦于**曼德勃罗集 (Mandelbrot set)** 的高效生成算法。\n\n## 曼德勃罗集的迭代公式\n\n核心的迭代公式非常简单，但却能产生无限复杂的边界：\n$Z_{n+1} = Z_n^2 + C$\n\n### 优化策略\n1. **逃逸时间算法优化**：使用平滑着色方案代替硬性分带。\n2. **多线程计算**：使用 Web Workers 在浏览器中进行并行渲染。\n3. **色彩映射**：使用了基于 HSL 的平滑过渡。\n\n希望这份档案能对后续参赛的同学有所启发。" }) });

  // 4. Invalidate cache
  await fetch(`${BASE_URL}/homepage/invalidate`, { method: 'POST', headers });

  console.log('Seeding completed successfully!');
}

seed().catch(console.error);
