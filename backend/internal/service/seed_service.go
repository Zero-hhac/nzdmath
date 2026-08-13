package service

import (
	"log/slog"
	"math-top/internal/model"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SeedService struct {
	db *gorm.DB
}

func NewSeedService(db *gorm.DB) *SeedService {
	return &SeedService{db: db}
}

func (s *SeedService) AutoSeed() {
	var count int64
	if err := s.db.Model(&model.News{}).Count(&count).Error; err != nil {
		slog.Error("检查数据库数据失败", "err", err)
		return
	}
	if count > 0 {
		slog.Info("检测到数据库已有数据，跳过自动 Seed")
		return
	}

	slog.Info("数据库中无数据，开始自动注入初始测试数据...")

	// 1. 创建普通测试会员用户
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	users := []model.User{
		{
			Username:     "testuser",
			PasswordHash: string(hashedPassword),
			Nickname:     "测试会员1号",
			Email:        stringPtr("test1@math-top.com"),
			Avatar:       "/uploads/avatars/2026/06/u12_1780838117858915000.png",
			RealName:     "张测试",
			ClassName:    "数学与应用数学2401班",
			Department:   "组织部",
			Status:       1,
			Bio:          "爱数学，爱生活。",
			Role:         "member",
		},
		{
			Username:     "math_geek",
			PasswordHash: string(hashedPassword),
			Nickname:     "几何狂热者",
			Email:        stringPtr("geek@math-top.com"),
			Avatar:       "/uploads/avatars/2026/06/u13_1780837528453519000.jpg",
			RealName:     "李几何",
			ClassName:    "信息与计算科学2402班",
			Department:   "宣传部",
			Status:       1,
			Bio:          "专注于拓扑学与分形几何研究。",
			Role:         "member",
		},
		{
			Username:     "EulerFan",
			PasswordHash: string(hashedPassword),
			Nickname:     "欧拉粉丝",
			Email:        stringPtr("euler@math-top.com"),
			Avatar:       "/uploads/avatars/2026/06/u14_1780839290521389000.png",
			RealName:     "王欧拉",
			ClassName:    "统计学2403班",
			Department:   "办公室",
			Status:       1,
			Bio:          "E^i*pi + 1 = 0.",
			Role:         "member",
		},
	}
	for i := range users {
		if err := s.db.Create(&users[i]).Error; err != nil {
			slog.Error("创建测试用户失败", "username", users[i].Username, "err", err)
		}
	}

	// 2. 创建资讯新闻
	news := []model.News{
		{
			Title:      "关于黎曼猜想的非正式讨论与阅读清单",
			Summary:    "本周末，数爱会将举办一场针对黎曼猜想的交流茶话会，附录有供大家提前预习的参考书单。",
			Category:   "讨论",
			Tag:        "黎曼猜想",
			Content:    "# 黎曼猜想 (Riemann Hypothesis) 阅读清单\n\n> 黎曼猜想是数学中最重要的未解问题之一。它不仅涉及素数的分布规律，也深深影响着现代数论的脉络。\n\n以下是我们将要在讨论会上分享和探讨的核心议题：\n\n## 1. 黎曼 Zeta 函数\n\n黎曼 Zeta 函数在复平面上的定义为：\n对于实部大于 1 的复数 s，定义为 $\\zeta(s) = \\sum_{n=1}^{\\infty} \\frac{1}{n^s}$。\n\n## 2. 核心猜想陈述\n\n**黎曼猜想断言：** 黎曼 Zeta 函数的所有非平凡零点，其解析延拓后的实部均为 `1/2`。\n这意味着，所有非平凡零点都落在复平面的**临界线**上。\n\n### 推荐阅读书目\n- 《素数之恋》 (Prime Obsession) - John Derbyshire\n- 《黎曼猜想漫谈》 - 卢昌海\n\n### 讨论会安排\n* **时间**：本周五下午 3:00 - 5:00\n* **地点**：图书馆 3 层数学沙龙区\n* **要求**：请大家提前阅读上述书目的第一章。\n\n我们非常期待你的加入，一起探讨这个迷人而深邃的话题！",
			Status:     1,
			IsFeatured: true,
			CoverURL:   "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
		{
			Title:      "计算拓扑学前沿讲座：从理论到代码",
			Summary:    "我们将介绍单纯复形、持久同调等概念，并使用 Python 进行现场演示。",
			Category:   "讲座",
			Tag:        "拓扑数据分析",
			Content:    "# 计算拓扑学：持久同调入门\n\n在数据科学 and 机器学习中，提取高维数据的形状特征变得越来越重要。这就是**计算拓扑学**大显身手的地方。\n\n## 什么是持久同调 (Persistent Homology)？\n\n它是拓扑数据分析 (TDA) 中的核心工具，通过在不同的尺度下构建拓扑结构（如单纯复形），来观察哪些拓扑特征（孔洞、连通分支）是短暂的噪音，哪些是“持久”的本质结构。\n\n### 代码演示预告\n\n在讲座中，我们将使用 Python 的 `Giotto-TDA` 和 `Ripser` 库来进行演示：\n\n```python\nimport numpy as np\nfrom ripser import ripser\nfrom persim import plot_diagrams\n\n# 生成一个圆上的噪声数据\ndata = np.random.random((100, 2))\n\n# 计算持久同调\ndiagrams = ripser(data)['dgms']\n\n# 绘制条形图\nplot_diagrams(diagrams, show=True)\n```\n\n我们现场见！",
			Status:     1,
			IsFeatured: true,
			CoverURL:   "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
	}
	for i := range news {
		if err := s.db.Create(&news[i]).Error; err != nil {
			slog.Error("创建资讯失败", "title", news[i].Title, "err", err)
		}
	}

	// 3. 创建学术活动 (Events)
	now := time.Now()
	events := []model.Event{
		{
			Title:      "拓扑数据分析 (TDA) 讨论班第一期",
			Summary:    "探讨计算拓扑中的单纯复形、过滤（Filtration）和持久贝蒂数计算。",
			Category:   "讲座",
			Content:    "# TDA 研讨班大纲\n\n### 讨论专题\n1. **Čech 复形与 Vietoris-Rips 复形** 的几何定义与差异。\n2. **持久图 (Persistence Diagram)** 与 **持久条形图 (Persistence Barcode)** 的构造及其稳定性定理。\n3. **瓶颈距离 (Bottleneck Distance)** 的度量定义与应用。\n\n请与会人员携带手提电脑，我们将现场运行实例代码。",
			Location:   "数学院报告厅 A102",
			StartTime:  now.AddDate(0, 0, 2),
			EndTime:    now.AddDate(0, 0, 2).Add(3 * time.Hour),
			Status:     1,
			IsFeatured: true,
			CoverUrl:   "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
		{
			Title:      "黎曼流形与微分几何研讨会",
			Summary:    "面向高年级本科生和研究生的微分几何系列研讨会，介绍曲率张量与测地线理论。",
			Category:   "研讨会",
			Content:    "# 微分几何研讨会大纲\n\n## 专题目录\n- 联络与共变导数 (Connection and Covariant Derivative)\n- 黎曼曲率张量与 Ricci 曲率张量\n- 雅可比场 (Jacobi Fields) 与第二变分公式\n\n欢迎所有感兴趣的同学参与！",
			Location:   "理科楼 301 教室",
			StartTime:  now.AddDate(0, 0, 5),
			EndTime:    now.AddDate(0, 0, 5).Add(2 * time.Hour),
			Status:     1,
			IsFeatured: false,
			CoverUrl:   "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
	}
	for i := range events {
		if err := s.db.Create(&events[i]).Error; err != nil {
			slog.Error("创建活动失败", "title", events[i].Title, "err", err)
		}
	}

	// 4. 创建作品档案 (Showcases)
	showcases := []model.Showcase{
		{
			Title:       "分形几何的可视化应用",
			Author:      "刘洋",
			Field:       "几何学",
			Competition: "全国计算机应用大赛",
			Summary:     "# 分形几何的无限嵌套\n\n分形之美在于其自相似性。这段作品主要聚焦于**曼德勃罗集 (Mandelbrot set)** 的高效生成算法。\n\n## 曼德勃罗集的迭代公式\n\n核心的迭代公式非常简单，但却能产生无限复杂的边界：\n$Z_{n+1} = Z_n^2 + C$\n\n### 优化策略\n1. **逃逸时间算法优化**：使用平滑着色方案代替硬性分带。\n2. **多线程计算**：使用 Web Workers 在浏览器中进行并行渲染。\n3. **色彩映射**：使用了基于 HSL 的平滑过渡。\n\n希望这份档案能对后续参赛的同学有所启发。",
			Status:      1,
			CoverURL:    "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
		{
			Title:       "第四届理事会竞选大会 PPT 演示",
			Author:      "数学爱好者协会",
			Field:       "组织活动",
			Competition: "理事会换届",
			Summary:     "# 第四届理事会竞选大会\n\n数学爱好者协会第四届理事会竞选大会的 HTML5 在线演示，使用了 Reveal.js 进行现代化转场渲染，并具备毛玻璃拟态的学术风格。\n\n## 大会主题\n- **竞选原则**：公开性原则、公平性原则\n- **开放竞选岗位**：会长 (1名)、副会长 (2名)、四部正副部长 (各1名)\n- **竞选评分权重**：现场投票 (70%) + 老师评审 (30%)\n\n请在下方点击“在线播放演示”查看全屏 PPT 效果！",
			Status:      1,
			CoverURL:    "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
			H5URL:       "/uploads/h5_unified_light/index.html",
		},
	}
	for i := range showcases {
		if err := s.db.Create(&showcases[i]).Error; err != nil {
			slog.Error("创建作品档案失败", "title", showcases[i].Title, "err", err)
		}
	}

	// 5. 创建下载资源 (Resources)
	resources := []model.Resource{
		{
			Title:         "数学协会2026年会员注册任务书",
			Summary:       "包含协会年度活动规划、会员责任义务说明、以及需要填写的个人详细学术背景登记表。",
			FileName:      "1780659025890581000_任务书.docx",
			FilePath:      "/uploads/resources/2026/06/1780659025890581000_任务书.docx",
			FileSize:      14713,
			FileType:      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			FileExt:       "docx",
			DownloadCount: 24,
			Status:        1,
			CoverURL:      "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
		{
			Title:         "25级数学协会会员名单及分组",
			Summary:       "2025级数学协会全体新会员名册、指导老师分组以及各个研究方向的详细排班表。",
			FileName:      "1780658975059338000_25级数学协会会员名单.xlsx",
			FilePath:      "/uploads/resources/2026/06/1780658975059338000_25级数学协会会员名单.xlsx",
			FileSize:      9512,
			FileType:      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			FileExt:       "xlsx",
			DownloadCount: 56,
			Status:        1,
			CoverURL:      "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
		{
			Title:         "数学专业优秀求职与学术简历模板",
			Summary:       "为数学系方向量身定制的 Go 语言工程、数据分析及学术研究型优质简历参考样例（含数学竞赛、建模荣誉描述等排版）。",
			FileName:      "1780658931442036000_go简历.pdf",
			FilePath:      "/uploads/resources/2026/06/1780658931442036000_go简历.pdf",
			FileSize:      54289,
			FileType:      "application/pdf",
			FileExt:       "pdf",
			DownloadCount: 108,
			Status:        1,
			CoverURL:      "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
		{
			Title:         "分形曼德勃罗集测试示例图",
			Summary:       "极具对称美学的分形 Mandelbrot 渲染图片资源，可用作电脑壁纸或排版背景配图。",
			FileName:      "1780658889292989000_奶龙.png",
			FilePath:      "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
			FileSize:      85312,
			FileType:      "image/png",
			FileExt:       "png",
			DownloadCount: 12,
			Status:        1,
			CoverURL:      "/uploads/resources/2026/06/1780658889292989000_奶龙.png",
		},
	}
	for i := range resources {
		if err := s.db.Create(&resources[i]).Error; err != nil {
			slog.Error("创建下载资源失败", "title", resources[i].Title, "err", err)
		}
	}

	// 6. 自动填充前 6 天的历史流量指标以呈现 Recharts 图表
	dMetrics := []model.DailyMetric{
		{Date: now.AddDate(0, 0, -6).Format("2006-01-02"), PV: 120, UV: 34, DAU: 12, CreatedAt: now.AddDate(0, 0, -6)},
		{Date: now.AddDate(0, 0, -5).Format("2006-01-02"), PV: 145, UV: 42, DAU: 18, CreatedAt: now.AddDate(0, 0, -5)},
		{Date: now.AddDate(0, 0, -4).Format("2006-01-02"), PV: 98, UV: 29, DAU: 11, CreatedAt: now.AddDate(0, 0, -4)},
		{Date: now.AddDate(0, 0, -3).Format("2006-01-02"), PV: 180, UV: 56, DAU: 25, CreatedAt: now.AddDate(0, 0, -3)},
		{Date: now.AddDate(0, 0, -2).Format("2006-01-02"), PV: 210, UV: 65, DAU: 32, CreatedAt: now.AddDate(0, 0, -2)},
		{Date: now.AddDate(0, 0, -1).Format("2006-01-02"), PV: 195, UV: 58, DAU: 28, CreatedAt: now.AddDate(0, 0, -1)},
	}
	for i := range dMetrics {
		s.db.Create(&dMetrics[i])
	}

	slog.Info("自动注入初始测试数据成功！")
}

func stringPtr(s string) *string {
	return &s
}
