<?php

namespace ZetaSkin;

class StaticMarkup
{
    public static function toc(array $dataToc): string
    {
        $sections = $dataToc['array-sections'] ?? [];
        if (! $sections) {
            return '';
        }

        $items = '';
        foreach ($sections as $s) {
            $items .= '<li class="m-0">'.self::tocNode($s, 0).'</li>';
        }

        return '<div class="ce-static"><div class="inline-block rounded w-full border border-[#9995] p-3 my-3"><nav>'
            .'<ul class="text-sm tracking-tight list-none m-0 p-0 text-foreground/80">'
            .'<li class="m-0 mb-3 flex items-center gap-2 font-bold"><span>목차</span></li> '
            .$items
            .'</ul></nav></div></div>';
    }

    private static function tocNode(array $s, int $depth): string
    {
        $anchor = (string) ($s['anchor'] ?? '');
        $label = trim(preg_replace('/<\/?[^>]+>/i', ' ', (string) ($s['line'] ?? '')));
        $html = '<a href="#'.self::e($anchor).'" class="flex w-full items-start gap-1 hover:no-underline text-foreground/80"'
            .' style="padding-left: calc('.($depth + 1).' * 0.75rem);">'
            .'<span class="shrink-0 text-muted-foreground/70"><span>'.self::e((string) ($s['number'] ?? '')).'</span>'
            .($depth === 0 ? ' <span>.</span>' : '').'</span> '
            .'<span class="flex-1 hover:underline">'.self::e(html_entity_decode($label, ENT_QUOTES | ENT_HTML5)).'</span></a>';

        $children = $s['array-sections'] ?? [];
        if ($children) {
            $html .= ' <ul class="p-0 list-none" role="list">';
            foreach ($children as $c) {
                $html .= '<li class="m-0">'.self::tocNode($c, $depth + 1).'</li>';
            }
            $html .= '</ul>';
        }

        return $html;
    }

    public static function disambig(?array $disambig, ?array $registration, int $articleId): string
    {
        $nodes = $disambig['nodes'] ?? [];
        if (! $nodes || ! $registration) {
            return '';
        }

        $items = '';
        foreach ($nodes as $node) {
            $text = self::e((string) ($node['text'] ?? ''));
            $link = (int) ($node['id'] ?? 0) === $articleId
                ? '<span class="font-semibold" aria-current="page">'.$text.'</span>'
                : '<a'.(! empty($node['new']) ? ' class="new"' : '').' href="'.self::e((string) ($node['href'] ?? '')).'">'.$text.'</a>';
            $desc = ! empty($node['description'])
                ? ' <span class="text-[0.6rem] text-muted-foreground">'.self::e((string) $node['description']).'</span>'
                : '';
            $items .= '<li class="w-full min-w-0"><div class="inline-flex min-w-0 flex-1 flex-wrap items-baseline gap-x-2 px-1.5 py-0.5">'
                .'다른 뜻에 대해서는 '.$link.$desc.' 문서를 참조하십시오.</div></li>';
        }

        return '<div class="ce-static"><aside class="my-3 w-fit max-w-full text-[0.8rem] rounded-md p-2 pl-5"'
            .' aria-label="'.self::e($registration['baseTitle'].' 동음이의 문서').'">'
            .'<div class="flex items-start gap-1">'
            .'<span class="inline-flex flex-none rounded p-0.5"><img class="size-5" src="https://upload.wikimedia.org/wikipedia/commons/4/4a/Disambig_grey.svg" alt="" aria-hidden="true"></span> '
            .'<ul class="m-0 flex min-w-0 flex-1 list-none flex-col gap-0.5 p-0">'.$items.'</ul>'
            .'</div></aside></div>';
    }

    private static function e(string $s): string
    {
        return htmlspecialchars($s, ENT_QUOTES);
    }
}
