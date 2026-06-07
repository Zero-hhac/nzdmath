import fs from 'fs';
import path from 'path';

const BASE_URL = 'http://localhost:8080/api/v1/admin';

async function seedExtra() {
  console.log('Logging in as admin...');
  const loginRes = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'admin123' })
  });
  
  if (!loginRes.ok) {
     console.error('Login failed'); return;
  }
  const { data: { token } } = await loginRes.json();
  
  console.log('Seeding extra markdown data...');
  const headers = { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` };

  // 1. Post News: 换届公告
  await fetch(`${BASE_URL}/news`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      title: '广西农业职业技术大学数学爱好者协会换届大会公告',
      summary: '旧岁已展千重锦，新年再进百尺竿。新一届数爱会换届选举工作即将正式启动，欢迎有志之士报名竞选！',
      category: '通知公告',
      tag: '协会换届',
      status: 1,
      is_featured: true,
      content: `
# 广西农业职业技术大学数学爱好者协会换届大会公告

亲爱的各位会员：

光阴荏苒，岁月如梭。在过去的一年里，**数学爱好者协会（数爱会）**在校团委的悉心指导下，在全体会员的共同努力下，成功举办了多场具有影响力的学术沙龙、竞赛培训和经验交流活动。

为保证协会工作的连续性，并为协会注入新鲜血液，经协会核心干部层商议，决定启动**新一届换届选举工作**。

## 竞选部门及岗位

本次换届将面向全体正式会员，公开竞选以下部门的核心骨干（部长及副部长）：

1. **办公室**：负责协会内部的日常统筹、档案管理、会议记录及财务收支统计。
2. **宣传部**：负责协会活动的线上线下宣传、海报设计、文案撰写及公众号运营。
3. **组织部**：主要负责筹划并执行各项讲座、培训、常规讨论会以及竞赛后勤支持。
4. **外联部**：负责拉取校内外赞助、与其他社团及高校数学组织的联谊与资源对接。

## 报名与时间安排

* **报名时间**：即日起至下周五晚 20:00
* **报名方式**：请前往“会员专区”下载并填写《换届竞选报名表》，以 \`姓名-竞选部门-意向职位\` 为邮件名，发送至协会官方邮箱。
* **换届大会（竞选答辩）时间**：预计于本月下旬在大学生活动中心召开（具体时间另行通知）。

> [!IMPORTANT]
> **换届寄语**
>
> 换届不仅是职务的交接，更是协会精神的传承。我们期待有责任心、对数学充满热情的你加入管理团队，**一起追踪每一个逻辑步长，一起做有目标、有方向的学习型协会**，为我校增添一份新的社团风采！

欢迎大家踊跃报名！
      `.trim()
    })
  });

  // 2. Post Event: 数学建模竞赛
  await fetch(`${BASE_URL}/events`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      title: '全国大学生数学建模竞赛 (CUMCM) 备战启动仪式与组队指南',
      summary: '一年一度的国赛即将拉开帷幕，我们将在此次启动仪式上解读最新赛规，并提供组队双选会。',
      category: '竞赛',
      location: '大学生活动中心报告厅',
      start_time: new Date(Date.now() + 86400000 * 5).toISOString(),
      end_time: new Date(Date.now() + 86400000 * 5 + 7200000).toISOString(),
      status: 1,
      is_featured: true,
      content: `
# 全国大学生数学建模竞赛 (CUMCM) 备战启动

全国大学生数学建模竞赛（CUMCM）是全国规模最大的基础性学科竞赛之一。对于锻炼思维、提升代码实操和论文撰写能力有着极大的帮助。

## 启动仪式核心议程

为了帮助我校同学更好地备战今年的国赛，协会特地组织了本次备战启动仪式：

1. **赛规解读与趋势分析**
   由指导老师剖析近年来 A、B、C 题的命题趋势（偏微分方程、运筹优化、大数据挖掘等）。
2. **国奖学长学姐经验分享**
   邀请了去年获得国家一等奖的团队，分享他们在三天三夜高压环境下的“极限生存”与分工策略。
3. **现场组队“双选会”**

### 组队建议矩阵

一个优秀的数模团队通常需要以下三类角色的完美配合：

| 角色定位 | 核心职责 | 推荐技能树 |
| :--- | :--- | :--- |
| **建模手** | 问题分析、建立核心数学模型 | 高等数学、运筹学、机器学习、统筹规划能力 |
| **编程手** | 算法实现、数据清洗、求解模型 | Python (Pandas/Scipy)、MATLAB、C++ |
| **论文手** | 论文撰写、数据可视化、排版 | LaTeX、Word 排版、图表绘制 (Matplotlib/Origin)、文字表达 |

> [!TIP]
> 还没有队友的同学不要担心，务必参加本次启动仪式，现场将有大量寻找“天选队友”的组队环节！

请各位有志于参赛的同学准时出席！
      `.trim()
    })
  });

  // 3. Post Event: 大学生数学竞赛 (CMC)
  await fetch(`${BASE_URL}/events`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      title: '全国大学生数学竞赛 (CMC) 考点冲刺串讲',
      summary: '针对即将在下个月举行的全国大学生数学竞赛，我们将进行一次高数核心考点的串讲和真题演练。',
      category: '讲座',
      location: '教四楼 205 阶梯教室',
      start_time: new Date(Date.now() + 86400000 * 7).toISOString(),
      end_time: new Date(Date.now() + 86400000 * 7 + 10800000).toISOString(),
      status: 1,
      is_featured: false,
      content: `
# 全国大学生数学竞赛 (CMC) 冲刺串讲

全国大学生数学竞赛 (CMC) 是一项旨在培养人才、服务教学、促进高校数学课程建设的顶尖赛事。

距离今年的初赛还有不到一个月的时间，为了帮助大家查漏补缺，协会学术部精心准备了本次冲刺串讲活动。

## 串讲内容大纲

本次串讲主要针对**非数学类**考纲中的核心难点：

### 1. 极限与连续
* 泰勒公式在极限计算中的高级应用
* 中值定理证明题的构造辅助函数技巧

### 2. 多元函数微积分
* 重积分的坐标系变换策略（极坐标、柱面坐标、球面坐标）
* 曲线与曲面积分的核心定理考法（格林公式、高斯公式、斯托克斯公式）

### 3. 级数
* 级数敛散性判别法的综合运用
* 傅里叶级数展开及其在级数求和中的应用

## 经典真题赏析
串讲最后，我们将共同推演几道近三年的高压压轴题。例如：
对于任意正整数 $n$，存在一个含有 $n$ 个连续合数的序列，这一经典结论在竞赛中可能以何种变体出现？

带上你的草稿纸和笔，让我们一起在题海中寻找数学的优雅之美！
      `.trim()
    })
  });

  // 4. Invalidate cache
  await fetch(`${BASE_URL}/homepage/invalidate`, { method: 'POST', headers });

  console.log('Extra seeding completed successfully!');
}

seedExtra().catch(console.error);
