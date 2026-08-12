import React from 'react';
import type { OverlayConfig } from '@/src/types/app';

interface FooterProps {
  openOverlay: (config: OverlayConfig) => void;
}

export const Footer: React.FC<FooterProps> = ({ openOverlay }) => {
  const footerLinks = [
    {
      label: '隐私政策',
      title: '隐私政策',
      body: '我们仅收集参与活动、下载资源与会员服务所需的最少信息，所有学术数据默认仅在协会内部授权范围内流转，并支持用户随时申请导出或删除个人资料。',
    },
    {
      label: '使用条款',
      title: '使用条款',
      body: '协会平台面向校内外数学学习者开放，资源使用须遵守原作者授权要求，禁止二次售卖、恶意爬取和篡改投稿内容；参与活动与投稿即视为同意平台规则。',
    },
    {
      label: '联系支持',
      title: '联系支持',
      body: '支持通道已覆盖资源反馈、活动咨询、会员申请和后台协作。你可以通过 `math_assoc@uni.edu`、教学楼 A 区 3 楼值班台，或会员专区内的工单入口联系协会。',
    },
    {
      label: '学术诚信',
      title: '学术诚信公约',
      body: '所有投稿、讲稿、竞赛材料与研究摘要均需注明来源。协会鼓励开放共享，但反对抄袭、代写和伪造数据，违规内容将被撤下并记录。',
    },
  ];

  return (
    <footer className="w-full bg-surface/70 backdrop-blur-lg border-t border-border">
      <div className="flex flex-col md:flex-row justify-between items-center py-10 px-6 md:px-10 gap-7 max-w-6xl mx-auto">
        <div className="text-lg text-charcoal font-bold tracking-tight">数学协会</div>
        
        <div className="text-sm text-zinc-500 text-center md:text-left font-medium">
          © {new Date().getFullYear()} 数学爱好者协会 · 探索逻辑之律
        </div>
        
        <div className="flex flex-wrap justify-center gap-x-5 gap-y-3">
          {footerLinks.map((item) => (
            <button
              key={item.label}
              onClick={() =>
                openOverlay({
                  title: item.title,
                  subtitle: '平台说明',
                  content: (
                    <div className="space-y-5">
                      <div className="rounded-[1.75rem] border border-border bg-canvas-alt p-6">
                        <p className="text-sm font-medium leading-7 text-zinc-600">{item.body}</p>
                      </div>
                      <div className="grid gap-4 md:grid-cols-3">
                        {['信息透明', '授权清晰', '可追溯记录'].map((point) => (
                          <div key={point} className="glass-card rounded-[1.5rem] p-5 text-center">
                            <div className="mb-2 text-xs font-bold uppercase tracking-[0.24em] text-zinc-400">Policy</div>
                            <div className="font-serif text-lg text-primary">{point}</div>
                          </div>
                        ))}
                      </div>
                    </div>
                  ),
                  actions: [{ label: '我已了解' }],
                })
              }
              className="text-xs text-zinc-600 hover:text-primary transition-colors"
            >
              {item.label}
            </button>
          ))}
        </div>
      </div>
    </footer>
  );
};
