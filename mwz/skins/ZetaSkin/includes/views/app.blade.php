<div id="app" class="min-h-screen">
    <nav id="navbar" class="w-full text-sm border-b z-40 bg-white dark:bg-slate-700">
        <div class="max-w-8xl mx-auto w-full md:flex whitespace-nowrap">
            <div class="inline-block md:flex-none">
                <div>
                    <div class="flex h-12">
                        <a href="/" class="navbar-link">
                            <img alt="zetawiki" src="/zeta.svg" width="12" height="16">
                            <span class="hidden lg:inline ml-2 text-lg font-semibold">제타위키</span></a>
                        </a>
                        <a href="/" class="navbar-link !text-yellow-700 dark:!text-yellow-200">위키</a>
                        <a href="/forum" class="navbar-link">포럼</a>
                        <a href="/tool/common-report" class="navbar-link">도구</a>
                    </div>
                </div>
            </div>
            <navbar-user-menu :user-menu='@json($userMenu)'></navbar-user-menu>
            <div class="flex-1">
                <navbar-search></navbar-search>
            </div>
        </div>
    </nav>
    <div class="max-w-8xl mx-auto">
        <div class="flex flex-row {{ $hasBinders ? 'has-binders' : '' }}">
            @if ($hasBinders)
                <div class="hidden md:block flex-none w-60">
                    <the-binder></the-binder>
                </div>
            @endif
            <div class="flex-auto">
                <div class="p-4 py-8 md:m-4 md:p-8 md:border md:rounded-md bg-white dark:bg-neutral-900">
                    <div class="text-3xl font-semibold">{!! $html_title !!}</div>
                    <div>{!! $html_subtitle !!}</div>
                    <div class="flow-root mt-2 text-sm">
                        <div class="float-left">
                            @if ($is_article && $action == 'view')
                                <page-meta historyhref="{{ $pageMenu['history']['href'] }}"></page-meta>
                            @endif
                        </div>
                        <div class="flex float-right">
                            @foreach ($pageBtns as $k => $v)
                                @if ($k != $action)
                                    @if ($k == 'edit')
                                        <a id="ca-edit" class="page-btn"
                                            href="{{ $v['href'] }}">{{ $v['text'] }}</a>
                                    @else
                                        <a class="page-btn"
                                            href="{{ $v['href'] }}">{{ $v['text'] == '읽기' ? '보기' : $v['text'] }}</a>
                                    @endif
                                @endif
                            @endforeach
                            <page-menu>
                                @foreach ($pageMenu as $k => $v)
                                    <li>
                                        <a href="{{ $v['href'] }}">{{ $v['text'] }}</a>
                                    </li>
                                @endforeach
                            </page-menu>
                        </div>
                    </div>
                    @isset($data_portlets['data-category-normal']['html-items'])
                        <div class="page-cats">
                            <ul>
                                {!! $data_portlets['data-category-normal']['html-items'] !!}
                            </ul>
                        </div>
                    @endisset
                    <article id="content" class="py-6" v-pre>
                        {!! $html_body_content !!}
                    </article>
                    @unless ($is_specialpage)
                        <the-runbox></the-runbox>
                        <page-foot></page-foot>
                    @endunless
                </div>
            </div>
            @unless ($is_specialpage)
                <div class="hidden hb-lg-md-block flex-none w-60">
                    <toc-main :datatoc='@json($data_toc)'></toc-main>
                </div>
            @endunless
        </div>
    </div>
    <layout-foot></layout-foot>
    <layout-remocon></layout-remocon>
</div>
<div id="box-app"></div>
