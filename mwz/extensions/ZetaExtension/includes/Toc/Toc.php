<?php
namespace ZetaExtension\Toc;

class Toc
{
    public static function onInternalParseBeforeLinks(&$parser, &$text)
    {
        if ($parser->getTitle()->getNamespace() != 0) {
            return true;
        }
        $text .= "__FORCETOC__";
        return true;
    }
}
