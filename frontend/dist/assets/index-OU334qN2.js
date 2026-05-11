const __vite__mapDeps=(i,m=__vite__mapDeps,d=(m.f||(m.f=["assets/Home-r68mvfa2.js","assets/problem-DQ1o-HLH.js","assets/element-plus-DlR6FFxp.js","assets/vendor-Cg51ANXx.js","assets/monaco-editor-By-2nyuq.js","assets/monaco-editor-X9GAyzG6.css","assets/Home-BmgW-x3j.css","assets/Login-B18u791-.js","assets/Login-CNYS0Hm9.css","assets/Register-BBMhlEUq.js","assets/Register-C0WnZqhW.css","assets/ProblemList-O6-swFg9.js","assets/problem-C1IQELmF.js","assets/ProblemList-CJY_QXla.css","assets/ProblemDetail-DOppDDvu.js","assets/submission-CwOwjGBY.js","assets/AIChat-C4rIxVB9.js","assets/AIChat-B7ln-qmQ.css","assets/ProblemDetail-qk9MVaQJ.css","assets/SubmissionStatus-xkuOrs5k.js","assets/SubmissionStatus-DeSS0IoS.css","assets/Profile-BJ3JicV4.js","assets/echarts-Bzpg1MfA.js","assets/Profile-DaNvSL7i.css","assets/AITraining-DPfru9mI.js","assets/AITraining-CU7_6D-T.css"])))=>i.map(i=>d[i]);
import{ax as Z,ay as Q,e as N,c as E,az as B,aA as tt,ag as p,H as A,L as P,P as h,M as i,O as c,Y as f,R as $,u as y,I as F,Z as j,F as et,X as nt,aB as ot,aC as at,au as rt,aD as st}from"./vendor-Cg51ANXx.js";import{E as it,a as lt,e as ct}from"./element-plus-DlR6FFxp.js";import{_ as v}from"./monaco-editor-By-2nyuq.js";(function(){const n=document.createElement("link").relList;if(n&&n.supports&&n.supports("modulepreload"))return;for(const a of document.querySelectorAll('link[rel="modulepreload"]'))o(a);new MutationObserver(a=>{for(const r of a)if(r.type==="childList")for(const s of r.addedNodes)s.tagName==="LINK"&&s.rel==="modulepreload"&&o(s)}).observe(document,{childList:!0,subtree:!0});function t(a){const r={};return a.integrity&&(r.integrity=a.integrity),a.referrerPolicy&&(r.referrerPolicy=a.referrerPolicy),a.crossOrigin==="use-credentials"?r.credentials="include":a.crossOrigin==="anonymous"?r.credentials="omit":r.credentials="same-origin",r}function o(a){if(a.ep)return;a.ep=!0;const r=t(a);fetch(a.href,r)}})();const V=Z.create({baseURL:"/api",timeout:15e3,headers:{"Content-Type":"application/json"}});V.interceptors.request.use(e=>{const n=localStorage.getItem("toj_token");return n&&(e.headers.Authorization=`Bearer ${n}`),e});V.interceptors.response.use(e=>e.data,e=>{var t,o,a;const n=((o=(t=e.response)==null?void 0:t.data)==null?void 0:o.message)||"请求失败，请稍后重试";return((a=e.response)==null?void 0:a.status)===401?(localStorage.removeItem("toj_token"),localStorage.removeItem("toj_user"),window.location.href="/login"):it.error(n),Promise.reject(e)});const m=(e=300)=>new Promise(n=>setTimeout(n,e+Math.random()*200)),d=e=>({code:0,message:"ok",data:e}),R=["动态规划","贪心","搜索","图论","数学","字符串","数据结构","模拟","排序","二分"],ut=["简单","中等","困难"];function mt(e=50){const n=[];for(let t=1;t<=e;t++){const o=ut[t%3];n.push({id:1e3+t,title:`${["两数之和","最长回文子串","合并区间","接雨水","全排列","最短路径","背包问题","编辑距离","岛屿数量","二叉树遍历"][t%10]} ${t>10?"II":""}`.trim(),difficulty:o,difficultyScore:o==="简单"?800+t%5*100:o==="中等"?1300+t%5*100:1800+t%5*100,tags:[R[t%10],R[(t+3)%10]],acceptRate:(40+Math.random()*50).toFixed(1),submitCount:Math.floor(100+Math.random()*5e3),accepted:t%4===0})}return n}const T=mt(),dt={content:`## 题目描述

给定一个整数数组 \`nums\` 和一个整数目标值 \`target\`，请你在该数组中找出和为目标值的两个整数，并返回它们的数组下标。

你可以假设每种输入只会对应一个答案，并且你不能使用两次相同的元素。

## 输入格式

第一行包含两个整数 $n$ 和 $target$，其中 $1 \\leq n \\leq 10^5$，$-10^9 \\leq target \\leq 10^9$。

第二行包含 $n$ 个整数 $a_1, a_2, \\ldots, a_n$，其中 $-10^9 \\leq a_i \\leq 10^9$。

## 输出格式

输出两个整数，表示和为 $target$ 的两个数的下标（从 0 开始），用空格分隔。

## 样例

### 输入
\`\`\`
4 9
2 7 11 15
\`\`\`

### 输出
\`\`\`
0 1
\`\`\`

## 提示

- 时间复杂度要求：$O(n)$
- 空间复杂度要求：$O(n)$

可以考虑使用哈希表来优化查找过程。`,timeLimit:1e3,memoryLimit:256,source:"TerminalOJ 原创题目"},C=["Accepted","Wrong Answer","Time Limit Exceeded","Runtime Error","Compilation Error","Pending"],J=["C++","Java","Python3","Go"];function pt(e=80){const n=[],t=Date.now();for(let o=0;o<e;o++){const a=C[Math.floor(Math.random()*C.length)];n.push({id:1e5+o,problemId:1e3+Math.floor(Math.random()*50)+1,problemTitle:T[Math.floor(Math.random()*50)].title,status:a,language:J[Math.floor(Math.random()*J.length)],runtime:a==="Accepted"?Math.floor(Math.random()*500)+10:null,memory:a==="Accepted"?(Math.random()*64+1).toFixed(1):null,createdAt:new Date(t-o*36e5*Math.random()*48).toISOString(),codeLength:Math.floor(Math.random()*2e3)+200})}return n.sort((o,a)=>new Date(a.createdAt)-new Date(o.createdAt))}const x=pt(),gt=[{id:1,title:"🎉 TerminalOJ 正式上线！",content:"欢迎使用 TerminalOJ 在线评测系统，祝大家刷题愉快！",date:"2026-04-01",type:"success"},{id:2,title:"📢 新增 AI 辅助训练功能",content:"现在你可以在做题时使用 AI 助手获取思路提示，同时支持独立的 AI 训练模式。",date:"2026-04-03",type:"info"},{id:3,title:"🔧 系统维护通知",content:"4月10日 02:00-04:00 将进行系统维护，届时评测服务暂停。",date:"2026-04-05",type:"warning"},{id:4,title:"🏆 每周竞赛开放报名",content:"第一期每周竞赛将于4月12日 19:00 开始，欢迎报名参加！",date:"2026-04-06",type:"primary"}],b={id:1,username:"coder_test",email:"test@terminaloj.com",avatar:"",bio:"热爱算法的开发者",rating:1520,rank:42,solvedCount:28,totalSubmissions:65,acceptRate:"43.1",registeredAt:"2026-03-15",solvedByDifficulty:{简单:15,中等:10,困难:3},solvedByAlgorithm:{动态规划:8,贪心:5,搜索:4,图论:3,数学:3,字符串:2,数据结构:2,模拟:1},recentActivity:[{date:"2026-04-06",count:3},{date:"2026-04-05",count:5},{date:"2026-04-04",count:2},{date:"2026-04-03",count:0},{date:"2026-04-02",count:4},{date:"2026-04-01",count:1}]},O={async login({username:e,password:n}){if(await m(500),!e||!n)throw new Error("请输入用户名和密码");return d({token:"mock_jwt_"+btoa(e)+"_"+Date.now(),user:{...b,username:e}})},async register({username:e,email:n,password:t}){if(await m(500),!e||!n||!t)throw new Error("请填写完整信息");return d({message:"注册成功"})},async getProfile(){return await m(300),d(b)},async updateProfile(e){return await m(300),d({...b,...e})},async getProblems({page:e=1,pageSize:n=20,keyword:t="",difficulty:o="",tag:a=""}={}){await m(400);let r=[...T];t&&(r=r.filter(l=>l.title.includes(t)||String(l.id).includes(t))),o&&(r=r.filter(l=>l.difficulty===o)),a&&(r=r.filter(l=>l.tags.includes(a)));const s=(e-1)*n;return d({list:r.slice(s,s+n),total:r.length})},async getProblemDetail(e){await m(300);const n=T.find(t=>t.id===Number(e));if(!n)throw new Error("题目不存在");return d({...n,...dt})},async submitCode({problemId:e,language:n,code:t}){await m(1500);const o=["Accepted","Wrong Answer","Time Limit Exceeded","Accepted","Accepted"],a=o[Math.floor(Math.random()*o.length)];return d({id:Date.now(),problemId:e,status:a,language:n,runtime:a==="Accepted"?Math.floor(Math.random()*200)+20:null,memory:a==="Accepted"?(Math.random()*32+2).toFixed(1):null,createdAt:new Date().toISOString()})},async getSubmissions({page:e=1,pageSize:n=20,problemId:t="",status:o="",sortBy:a="time"}={}){await m(400);let r=[...x];t&&(r=r.filter(l=>l.problemId===Number(t))),o&&(r=r.filter(l=>l.status===o)),a==="problemId"&&r.sort((l,w)=>l.problemId-w.problemId);const s=(e-1)*n;return d({list:r.slice(s,s+n),total:r.length})},async getSubmissionDetail(e){await m(200);const n=x.find(t=>t.id===Number(e));return d(n||null)},async aiChat({message:e,history:n,problem_id:t,conversation_id:o}){await m(800+Math.random()*1200);const a=t?`

> 当前关联题目 ID: ${t}`:"",r=[`这是一个很好的问题！让我来分析一下：

首先，我们需要理解问题的核心：

1. **分析输入输出**：仔细观察给定的样例
2. **选择合适的算法**：根据时间复杂度要求选择
3. **实现与优化**：编写代码并进行优化

你可以尝试使用 **哈希表** 来优化查找过程，时间复杂度为 $O(n)$。

\`\`\`cpp
unordered_map<int, int> mp;
for (int i = 0; i < n; i++) {
    if (mp.count(target - nums[i])) {
        return {mp[target - nums[i]], i};
    }
    mp[nums[i]] = i;
}
\`\`\`${a}`,`让我帮你理清思路：

这道题可以用 **动态规划** 来解决。

### 状态定义
设 $dp[i]$ 表示以第 $i$ 个元素结尾的最优解。

### 状态转移
$$dp[i] = \\max_{j < i}(dp[j] + w(j, i))$$

### 边界条件
- $dp[0] = 0$

### 复杂度分析
- 时间：$O(n^2)$，可以用数据结构优化到 $O(n \\log n)$
- 空间：$O(n)$${a}`,`好的，我来给你一些提示：

**关键观察**：这道题本质上是一个 **图论问题**。

1. 将每个元素看作图中的节点
2. 根据条件建边
3. 然后在图上进行 BFS/DFS

> 💡 提示：注意边界条件的处理，特别是当 $n = 1$ 的情况。

如果你需要更详细的解释，请告诉我具体哪个部分不理解。${a}`];return d({reply:r[Math.floor(Math.random()*r.length)],conversationId:o||"mock_conv_"+Date.now(),provider:"mock"})},async getAIHistory(){return await m(300),d({conversations:[]})},async getAIMessages(){return await m(200),d({conversation:null,messages:[]})},async aiCodeDiagnosis({problemId:e,language:n,code:t}){await m(700);const o=[];t!=null&&t.trim()||o.push({severity:"error",message:"代码为空",hint:"请先输入待诊断代码。"}),t!=null&&t.includes("TODO")&&o.push({severity:"warning",message:"代码中包含 TODO 占位",hint:"提交前补齐逻辑。"}),((t==null?void 0:t.match(/{/g))||[]).length!==((t==null?void 0:t.match(/}/g))||[]).length&&o.push({severity:"error",message:"花括号数量不匹配",hint:"检查代码块闭合。"}),o.length===0&&o.push({severity:"info",message:"Mock 检查未发现明显语法级问题",hint:"继续用边界用例验证。"});const a=`### 代码诊断

题目：#${e}，语言：${n}

#### 发现的问题

${o.map(r=>`- **${r.severity}**：${r.message}。${r.hint}`).join(`
`)}

#### 建议

- 先跑样例，再补充极值和重复数据。
- 根据题目约束重新核对时间复杂度。`;return d({summary:"Mock 代码诊断完成。",issues:o,suggestions:["补充边界用例","检查复杂度","确认输入输出格式"],rawMarkdown:a,provider:"mock"})},async aiKnowledgeGraph({problemId:e,scope:n="recent"}={}){await m(700);const t=[{id:"user",label:"当前用户",type:"user",weight:1},{id:"tag:动态规划",label:"动态规划",type:"algorithm",weight:8},{id:"tag:图论",label:"图论",type:"algorithm",weight:3},{id:"status:Accepted",label:"Accepted",type:"status",weight:12}];e&&t.push({id:`problem:${e}`,label:`题目 ${e}`,type:"problem",weight:1});const o=[{source:"user",target:"tag:动态规划",type:"strong_at",weight:8},{source:"user",target:"tag:图论",type:"need_practice",weight:3},{source:"user",target:"status:Accepted",type:"has_result",weight:12}],a=`### 学习知识图谱

已按 \`${n}\` 范围生成 Mock 图谱。

- 节点数：${t.length}
- 关系数：${o.length}

#### 建议

- 继续巩固动态规划的状态设计。
- 增加图论最短路和连通性题目练习。`;return d({summary:"Mock 知识图谱生成完成。",nodes:t,edges:o,rawMarkdown:a,provider:"mock"})},async aiSolve({problemId:e,question:n="",level:t="hint"}){return await m(800),d({answer:`### #${e} 解题辅助

当前级别：\`${t}\`。

先从暴力思路出发，确认状态或数据结构设计，再根据约束优化。${n?`

你的问题：

> ${n}`:""}`,hints:["手算样例观察规律","确认边界条件","写出复杂度再提交"],complexity:"Mock 模式下建议目标复杂度控制在题目约束可接受范围内。",provider:"mock"})},async getAnnouncements(){return await m(200),d(gt)}},S={login:e=>O.login(e),register:e=>O.register(e),getProfile:()=>O.getProfile(),updateProfile:e=>O.updateProfile(e)},U=Q("user",()=>{const e=N(localStorage.getItem("toj_token")||""),n=N(JSON.parse(localStorage.getItem("toj_user")||"null")),t=E(()=>!!e.value),o=E(()=>{var u;return((u=n.value)==null?void 0:u.username)||""});function a(u,g){e.value=u,n.value=g,localStorage.setItem("toj_token",u),localStorage.setItem("toj_user",JSON.stringify(g))}function r(){e.value="",n.value=null,localStorage.removeItem("toj_token"),localStorage.removeItem("toj_user")}async function s(u){const g=await S.login(u);return a(g.data.token,g.data.user),g}async function l(u){return await S.register(u)}async function w(){const u=await S.getProfile();return n.value=u.data,localStorage.setItem("toj_user",JSON.stringify(u.data)),u.data}async function _(u){const g=await S.updateProfile(u);return n.value={...n.value,...g.data},localStorage.setItem("toj_user",JSON.stringify(n.value)),g.data}function M(){r()}return{token:e,userInfo:n,isLoggedIn:t,username:o,login:s,register:l,logout:M,fetchProfile:w,updateProfile:_,setAuth:a}}),q=(e,n)=>{const t=e.__vccOpts||e;for(const[o,a]of n)t[o]=a;return t},_t={class:"navbar"},ft={class:"navbar-inner"},ht={class:"navbar-left"},yt={class:"nav-links"},vt={class:"navbar-right"},wt={class:"user-info"},$t={class:"username"},At={__name:"NavBar",setup(e){const n=B(),t=tt(),o=U();function a(r){r==="profile"?t.push("/profile"):r==="logout"&&(o.logout(),t.push("/login"))}return(r,s)=>{const l=p("router-link"),w=p("HomeFilled"),_=p("el-icon"),M=p("Document"),u=p("DataAnalysis"),g=p("MagicStick"),G=p("el-avatar"),z=p("ArrowDown"),W=p("User"),D=p("el-dropdown-item"),K=p("SwitchButton"),X=p("el-dropdown-menu"),Y=p("el-dropdown"),L=p("el-button");return A(),P("header",_t,[h("div",ft,[h("div",ht,[i(l,{to:"/",class:"logo"},{default:c(()=>[...s[0]||(s[0]=[h("span",{class:"logo-icon"},"⚡",-1),h("span",{class:"logo-text"},"TerminalOJ",-1)])]),_:1}),h("nav",yt,[i(l,{to:"/",class:$({active:y(n).name==="home"})},{default:c(()=>[i(_,null,{default:c(()=>[i(w)]),_:1}),s[1]||(s[1]=f("首页 ",-1))]),_:1},8,["class"]),i(l,{to:"/problems",class:$({active:y(n).name==="problems"})},{default:c(()=>[i(_,null,{default:c(()=>[i(M)]),_:1}),s[2]||(s[2]=f("题库 ",-1))]),_:1},8,["class"]),i(l,{to:"/status",class:$({active:y(n).name==="status"})},{default:c(()=>[i(_,null,{default:c(()=>[i(u)]),_:1}),s[3]||(s[3]=f("评测 ",-1))]),_:1},8,["class"]),i(l,{to:"/ai",class:$({active:y(n).name==="ai-training"})},{default:c(()=>[i(_,null,{default:c(()=>[i(g)]),_:1}),s[4]||(s[4]=f("AI 训练 ",-1))]),_:1},8,["class"])])]),h("div",vt,[y(o).isLoggedIn?(A(),F(Y,{key:0,trigger:"click",onCommand:a},{dropdown:c(()=>[i(X,null,{default:c(()=>[i(D,{command:"profile"},{default:c(()=>[i(_,null,{default:c(()=>[i(W)]),_:1}),s[5]||(s[5]=f("个人中心 ",-1))]),_:1}),i(D,{command:"logout",divided:""},{default:c(()=>[i(_,null,{default:c(()=>[i(K)]),_:1}),s[6]||(s[6]=f("退出登录 ",-1))]),_:1})]),_:1})]),default:c(()=>{var k;return[h("div",wt,[i(G,{size:32,src:((k=y(o).userInfo)==null?void 0:k.avatar)||void 0},{default:c(()=>[f(j(y(o).username.charAt(0).toUpperCase()),1)]),_:1},8,["src"]),h("span",$t,j(y(o).username),1),i(_,null,{default:c(()=>[i(z)]),_:1})])]}),_:1})):(A(),P(et,{key:1},[i(l,{to:"/login"},{default:c(()=>[i(L,{type:"primary",round:"",size:"small"},{default:c(()=>[...s[7]||(s[7]=[f("登录",-1)])]),_:1})]),_:1}),i(l,{to:"/register",style:{"margin-left":"8px"}},{default:c(()=>[i(L,{round:"",size:"small"},{default:c(()=>[...s[8]||(s[8]=[f("注册",-1)])]),_:1})]),_:1})],64))])])])}}},It=q(At,[["__scopeId","data-v-d5c09936"]]),Ot={id:"terminal-oj"},St={__name:"App",setup(e){const n=B(),t=E(()=>!["login","register"].includes(n.name));return(o,a)=>{const r=p("router-view");return A(),P("div",Ot,[t.value?(A(),F(It,{key:0})):nt("",!0),h("main",{class:$({"with-nav":t.value})},[i(r)],2)])}}},Mt=q(St,[["__scopeId","data-v-1faf3a2e"]]),bt=[{path:"/",name:"home",component:()=>v(()=>import("./Home-r68mvfa2.js"),__vite__mapDeps([0,1,2,3,4,5,6])),meta:{title:"首页 - TerminalOJ"}},{path:"/login",name:"login",component:()=>v(()=>import("./Login-B18u791-.js"),__vite__mapDeps([7,3,2,4,5,8])),meta:{title:"登录 - TerminalOJ",guest:!0}},{path:"/register",name:"register",component:()=>v(()=>import("./Register-BBMhlEUq.js"),__vite__mapDeps([9,3,2,4,5,10])),meta:{title:"注册 - TerminalOJ",guest:!0}},{path:"/problems",name:"problems",component:()=>v(()=>import("./ProblemList-O6-swFg9.js"),__vite__mapDeps([11,3,12,1,2,4,5,13])),meta:{title:"题目列表 - TerminalOJ"}},{path:"/problem/:id",name:"problem-detail",component:()=>v(()=>import("./ProblemDetail-DOppDDvu.js"),__vite__mapDeps([14,3,12,1,15,16,2,17,4,5,18])),meta:{title:"题目详情 - TerminalOJ",auth:!0}},{path:"/status",name:"status",component:()=>v(()=>import("./SubmissionStatus-xkuOrs5k.js"),__vite__mapDeps([19,15,3,2,4,5,20])),meta:{title:"评测状态 - TerminalOJ",auth:!0}},{path:"/profile",name:"profile",component:()=>v(()=>import("./Profile-BJ3JicV4.js"),__vite__mapDeps([21,2,3,22,4,5,23])),meta:{title:"个人中心 - TerminalOJ",auth:!0}},{path:"/ai",name:"ai-training",component:()=>v(()=>import("./AITraining-DPfru9mI.js"),__vite__mapDeps([24,16,3,2,17,4,5,25])),meta:{title:"AI 训练 - TerminalOJ",auth:!0}}],H=ot({history:at(),routes:bt});H.beforeEach((e,n,t)=>{document.title=e.meta.title||"TerminalOJ";const o=U();e.meta.auth&&!o.isLoggedIn?t({name:"login",query:{redirect:e.fullPath}}):e.meta.guest&&o.isLoggedIn?t({name:"home"}):t()});const I=rt(Mt);for(const[e,n]of Object.entries(lt))I.component(e,n);I.use(st());I.use(H);I.use(ct);I.mount("#app");export{q as _,O as m,U as u};
