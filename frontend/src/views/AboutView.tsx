import React from 'react';
import type { ViewProps } from '@/src/types/app';

const members = [
  { name: '张教授', role: '荣誉顾问', field: '微分几何', avatar: 'https://i.pravatar.cc/150?u=1' },
  { name: '李明', role: '协会会长', field: '解析数论', avatar: 'https://i.pravatar.cc/150?u=2' },
  { name: '王梦', role: '学术部长', field: '概率统计', avatar: 'https://i.pravatar.cc/150?u=3' },
  { name: '刘洋', role: '宣传部长', field: '计算数学', avatar: 'https://i.pravatar.cc/150?u=4' }
];

export const AboutView: React.FC<ViewProps> = ({ navigate, openOverlay }) => {
  return (
    <div className="space-y-20">
      {/* Intro Section */}
      <section className="flex flex-col md:flex-row items-center gap-12">
        <div className="flex-1 space-y-8">
          <div className="page-intro space-y-2">
            <div className="section-kicker">About</div>
            <h2 className="section-title">关于我们</h2>
            <p className="section-subtitle italic">交流思想、提高能力、团队协作、开拓创新</p>
          </div>
          <div className="space-y-6 text-soft-body leading-relaxed font-medium">
            <p>
              广西农业职业技术大学数学爱好者协会（简称数爱会）成立于 2022 年 11 月 2 日，是由广大数学爱好者组成的一个学生社团。
            </p>
            <p>
              我们通过对数学史、数学文化的交流，培养大家对科学探究的兴趣。协会设有办公室、宣传部、组织部、外联部四个核心部门。我们的目标是凝聚更多热爱数学建模与考研数学的同学，提升团队协作解决实际问题的能力，共同为大学增添一份新的社团风采！
            </p>
          </div>
          <div className="flex gap-4">
             <div className="text-center p-6 glass-card rounded-[2rem] flex-1">
                <div className="text-3xl font-bold text-primary mb-1">2022</div>
                <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest">年正式成立</div>
             </div>
             <div className="text-center p-6 glass-card rounded-[2rem] flex-1">
                <div className="text-3xl font-bold text-primary mb-1">4</div>
                <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest">大核心职能部门</div>
             </div>
             <div className="text-center p-6 glass-card rounded-[2rem] flex-1">
                <div className="text-3xl font-bold text-primary mb-1">100%</div>
                <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest">专注学术交流</div>
             </div>
          </div>
        </div>
        <div className="flex-shrink-0 w-full md:w-[400px] aspect-square rounded-[3rem] overflow-hidden glass-card p-0 border-none shadow-2xl rotate-3">
          <img 
            alt="Association Archive" 
            className="w-full h-full object-cover transition-transform duration-1000 hover:scale-105" 
            src="https://lh3.googleusercontent.com/aida-public/AB6AXuB2awq8xo2qo1S__BhSiPyCHbasml_p-jKCEosh6cGwEQrnWPM7t4BIYJ-eLehl9a973yhP9-M521B-hu4vt-IrUU3XH7n3A-JTTTW4XDMjozLeqwzXjT_qTNqxwVnHHS6KND2CnaSOmw_ipT9hY7Sk6rZhU8EXxMLAmRXyG5YEL-Gi4206EI4ue4x_Bq4rJVgEie7vOVyRIxdeO4QA57it8HXN4lMEd4jMlRiC4HgILk0_cvj3Qp1VNEdY2u8MIfTSXknqLcwTBM2b" 
          />
        </div>
      </section>

      {/* Team Section */}
      <section className="space-y-12">
        <div className="text-center">
           <h3 className="text-3xl font-serif text-primary">核心团队</h3>
           <p className="text-zinc-500 font-medium mt-2">引领理性的探索者</p>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
          {members.map((member, i) => (
            <div key={i} className="text-center space-y-4 group">
              <div className="w-24 h-24 mx-auto rounded-full overflow-hidden border-2 border-white shadow-xl glass-card p-0 group-hover:scale-110 transition-transform duration-500">
                <img src={member.avatar} alt={member.name} className="w-full h-full object-cover" />
              </div>
              <div>
                <h4 className="font-serif text-lg text-primary">{member.name}</h4>
                <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest mb-1">{member.role}</div>
                <div className="math-tag !bg-white group-hover:border-primary transition-colors">{member.field}</div>
              </div>
            </div>
          ))}
        </div>
      </section>
      
      {/* Contact Section */}
      <section className="sidebar-panel rounded-[3rem] p-12 md:p-20 text-center space-y-8 bg-gradient-to-br from-white/92 to-[#e3f2fd]/48">
        <h3 className="text-3xl font-serif text-primary">遇见同路人</h3>
        <p className="text-soft-body max-w-xl mx-auto font-medium leading-relaxed">
          无论你是初出茅庐的数学爱好者，还是深耕领域的专业学者，数学协会的大门始终为你敞开。加入我们，共赴真理之径。
        </p>
        <div className="flex flex-col sm:flex-row justify-center gap-4">
           <div className="flex flex-col items-center gap-2">
              <span className="material-symbols-outlined text-primary text-3xl">mail</span>
              <span className="text-xs font-bold text-zinc-500">math_assoc@uni.edu</span>
           </div>
           <div className="hidden sm:block w-px h-12 bg-zinc-200"></div>
           <div className="flex flex-col items-center gap-2">
              <span className="material-symbols-outlined text-primary text-3xl">share_location</span>
              <span className="text-xs font-bold text-zinc-500">教学楼 A 区 3 楼学术角</span>
           </div>
        </div>
        <button
          onClick={() =>
            openOverlay({
              title: '加入我们',
              subtitle: '协会入会说明',
              content: (
                <div className="space-y-5">
                  <div className="rounded-[1.75rem] border border-white/70 bg-[#f7f9fb] p-6">
                    <p className="text-sm font-medium leading-7 text-zinc-600">
                      我们欢迎基础阶段学习者、竞赛选手、研究型成员和跨学科合作伙伴加入。系统会根据你的兴趣方向，为你匹配活动、资料和社群入口。
                    </p>
                  </div>
                  <div className="grid gap-4 md:grid-cols-3">
                    {[
                      ['基础成长', '适合想系统提升数学能力的新成员。'],
                      ['竞赛突破', '适合建模、竞赛、专题训练方向。'],
                      ['研究协作', '适合导师制、论文与项目合作方向。'],
                    ].map(([label, text]) => (
                      <div key={label} className="glass-card rounded-[1.5rem] p-5">
                        <div className="mb-2 font-serif text-lg text-primary">{label}</div>
                        <p className="text-sm font-medium leading-7 text-zinc-600">{text}</p>
                      </div>
                    ))}
                  </div>
                </div>
              ),
              actions: [
                { label: '进入会员专区', onClick: () => navigate('portal') },
                { label: '查看近期活动', variant: 'secondary', onClick: () => navigate('events') },
              ],
            })
          }
          className="btn-primary px-10 py-4 text-sm uppercase tracking-widest mt-6"
        >
          立即加入我们
        </button>
      </section>
    </div>
  );
};
