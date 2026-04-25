package llm

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type PromptTemplate struct {
	SystemPrompt string `yaml:"system_prompt"`
	UserPrompt   string `yaml:"user_prompt"`
}

// LoadPromptTemplate 从 YAML 文件加载 Prompt 模板。
// 参数：path - 模板文件路径。
// 返回：PromptTemplate - 模板对象；error - 加载失败错误。
func LoadPromptTemplate(path string) (PromptTemplate, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return PromptTemplate{}, err
	}
	tpl := PromptTemplate{}
	if err = yaml.Unmarshal(content, &tpl); err != nil {
		return PromptTemplate{}, err
	}
	if strings.TrimSpace(tpl.UserPrompt) == "" {
		return PromptTemplate{}, fmt.Errorf("prompt user_prompt is empty")
	}
	return tpl, nil
}

// RenderFortunePrompt 渲染模板占位符为具体用户资料。
// 参数：tpl - 模板对象；profile - 运势输入资料。
// 返回：string - 渲染后的 Prompt 文本。
func RenderFortunePrompt(tpl PromptTemplate, profile FortuneProfile) string {
	focusPool := []string{
		// ================= 职场与学业篇 =================
		"跨部门沟通与甩锅扯皮",
		"处理突发性的棘手问题",
		"被临时加派任务的应对",
		"临门一脚却卡壳的项目推进",
		"上级突如其来的灵魂拷问",
		"连轴转的无效会议轰炸",
		"攻克一项困扰已久的技术或业务难题",
		"枯燥繁琐流程的自动化提效",
		"代码或方案评审中的意见交锋",
		"排查难以复现的幽灵 Bug 或历史遗留问题",
		"方案被推翻重来的巨大挫败感",
		"获得关键人物的资源支持与认可",
		"摸鱼被抓包或进度落后的心虚",
		"重构或优化现有工作流的强烈冲动",
		"与资深前辈的高效经验对齐",
		"沉浸式心流状态带来的极高产出",
		"面对模糊需求的艰难拆解与落地",
		"忙碌一天却毫无建树的瞎忙碌感",
		"顺利完成阶段性里程碑的松弛感",
		"零散知识库或笔记资料的系统性整理",
		"死线 (Deadline) 逼近时的极限肾上腺素爆发",
		"被琐碎支持性工作无情打断核心节奏",
		"跨团队信息同步延迟导致的信息差",
		"发现现有架构或规则中的致命漏洞",
		"周报或总结撰写时的词穷与强行包装",
		"灵光一闪的创新性极佳解法",
		"生产工具或开发环境崩溃带来的心态崩盘",
		"摸准团队隐性规则后的游刃有余",

		// ================= 人际与边界篇 =================
		"面对他人无理要求的隐忍",
		"与同事边界感的摩擦",
		"复杂人情局的站队压力",
		"面对家庭琐事的耐心消耗",
		"识破他人伪装后的看破不说破",
		"一场酣畅淋漓的深度对谈",
		"拒绝他人不合理请求时的道德绑架",
		"意外收到旧相识的联络或消息",
		"帮人收拾烂摊子带来的极度憋屈",
		"敏感捕捉到团队氛围的微妙变化",
		"与合拍伙伴的心有灵犀瞬间",
		"面对笨蛋队友的涵养大挑战",
		"线上文字沟通词不达意引发的误解",
		"参与无聊团建或社交应酬的被迫假笑",
		"关键时刻获得他人的隐蔽提点",
		"试图改变他人固执观念的徒劳感",
		"在尴尬群聊中做那个打破冷场的人",
		"远离内耗型人格的社交断舍离",
		"一次意外的赞美带来的巨大情绪回血",
		"发现自己莫名成了八卦漩涡的中心",

		// ================= 情绪与心理防御篇 =================
		"拖延与自我否定的拉扯",
		"短视频信息轰炸后的情绪疲劳",
		"突如其来的无意义感与空虚",
		"对未来某事的过度预期与提前焦虑",
		"完美主义作祟导致的不肯放过自己",
		"卸下社交伪装后的长叹与放空",
		"被生活微小确幸治愈的瞬间",
		"对抗周一或周五的特定生物钟倦怠",
		"嫉妒心或胜负欲的暗中发酵",
		"面对选择困难症时的抓狂与内耗",
		"一次彻底的情绪决堤与释放",
		"专注力被碎片信息严重切割",
		"不被理解时的习惯性闭嘴与沉默",
		"旧事重提或回忆杀带来的深夜 Emo",
		"成功洗脑自己接受现实的自我PUA",
		"找回久违的掌控感与强大自信",
		"面对不确定性时的盲目乐观",
		"压抑已久的表达欲突然倾泻爆发",
		"对某种特定事物或目标的执念升温",
		"强迫症得不到满足的百爪挠心",

		// ================= 财富与物质欲篇 =================
		"强烈的冲动消费欲望",
		"算计月度收支时的捉襟见肘感",
		"薅到羊毛或获得意外小财的狂喜",
		"为情绪价值买单的盲目氪金",
		"遗失物品或损坏财物的破财之灾",
		"面对促销活动与消费陷阱的理智博弈",
		"一笔延期款项或报销的终于到账",
		"理财账户波动的刺激体验",
		"舍不得扔旧物带来的物理空间挤压",
		"被强烈安利成功后的疯狂种草",
		"发现高性价比替代品的小确幸",
		"缴纳各类月费与账单时的肉痛",
		"为健康或自我提升的一笔咬牙投资",
		"试图开启副业搞钱的内心躁动",
		"被朋友请客或投喂的白嫖快乐",

		// ================= 健康与日常作息篇 =================
		"健康透支后的补救",
		"久坐伏案带来的肩颈僵硬严重警告",
		"报复性熬夜与第二天起不来的恶性循环",
		"突然觉醒的运动健身热血",
		"外卖吃腻后对新鲜食材的极端渴望",
		"睡眠质量极佳带来的精神饱满",
		"换季或温度变化引发的小病痛",
		"咖啡因免疫后的纯靠意志力硬撑",
		"极度渴望亲近自然或晒太阳补充能量",
		"突发奇想的居家大扫除与空间重置",
		"尝试一种新的作息或饮食规律",
		"屏幕盯太久导致的视觉与大脑双重模糊",
		"对某项高热量垃圾食品的难以抗拒",
		"成功抵御宵夜诱惑的自律大胜利",
		"察觉到身体某处长期的隐秘不适",
	}
	dailyFocus := ""
	if len(focusPool) > 0 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		pickCount := 3
		if len(focusPool) < pickCount {
			pickCount = len(focusPool)
		}
		indexes := rng.Perm(len(focusPool))[:pickCount]
		picked := make([]string, 0, pickCount)
		for _, idx := range indexes {
			picked = append(picked, focusPool[idx])
		}
		dailyFocus = strings.Join(picked, "、")
	}
	r := tpl.UserPrompt
	r = strings.ReplaceAll(r, "{birthday}", profile.Birthday)
	r = strings.ReplaceAll(r, "{today}", profile.Today)
	r = strings.ReplaceAll(r, "{constellation}", profile.Constellation)
	r = strings.ReplaceAll(r, "{gender}", profile.Gender)
	r = strings.ReplaceAll(r, "{city}", profile.City)
	r = strings.ReplaceAll(r, "{occupation}", profile.Occupation)
	r = strings.ReplaceAll(r, "{daily_focus}", dailyFocus)
	return r
}
