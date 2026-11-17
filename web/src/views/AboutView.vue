<script setup lang="ts">
import { VERSION_LINE } from '../version'
</script>

<template>
  <div class="container section">
    <article class="about-prose about-single">
      <div class="eyebrow">ABOUT</div>
      <h1>关于</h1>
      <p class="lead">这个站只放一类东西：我自己做完、踩过、最后改对了的开发记录。</p>

      <h2>我是谁？</h2>
      <p>
        我是 Mr lin，业余时间大多花在写小工具上。日常写 C# 和 Go，偶尔写点前端，做的东西基本都是自己用得上的：剪贴板管理、局域网传文件、Excel 批量合并、记账、自托管的邮件客户端。
      </p>
      <p>
        不是科班出身的程序员，很多结论都是边做边试出来的。所以这里的文章里经常出现「这条我没验证」「只测过 Gmail」这样的句子——那是实情，不是客套。
      </p>

      <h2>为什么建这个站？</h2>
      <p>
        因为自己踩坑的时候，搜到的教程大多讲得不清楚：要么默认你已经会了，要么把关键步骤略成一句「配置一下就行」。真正卡住的时候，缺的恰好就是那一句。
      </p>
      <p>
        后来发现，能把一件事讲清楚的前提是——刚踩完，细节还记得。所以每解决一个问题就写一篇：当时的判断、走过的错路、最后改成了什么、还有什么没解决。
      </p>
      <p>
        不追热点，也不写「三分钟精通」。文章里的报错、数字、命令，都是从日志和代码里抄下来的。
      </p>

      <h2>在哪里可以找到我？</h2>
      <p>
        文章对应项目的源码都放在 GitHub：<a
          href="https://github.com/mylxnet"
          target="_blank"
          rel="noopener"
          >github.com/mylxnet</a
        >。每篇文末都会标明它属于哪个项目，仓库是唯一的对外入口。
      </p>
      <p>本站不开放注册、评论与投稿。想讨论某一篇，直接在对应仓库里提 issue 更快。</p>

      <p class="about-version">{{ VERSION_LINE }}</p>
    </article>
  </div>
</template>
