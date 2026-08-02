<?php

namespace ZetaExtension\Disambig;

use MediaWiki\MediaWikiServices;

final class DisambigHooks
{
    public static function onMakeGlobalVariablesScript(array &$vars, $out): void
    {
        $vars['disambigRegistration'] = self::getRegistrationData($out);
    }

    public static function getRegistrationData($out): ?array
    {
        if ($out->getActionName() !== 'view') {
            return null;
        }

        $title = $out->getTitle();
        if (! $title || $title->getNamespace() !== NS_MAIN || $title->getArticleID() < 1) {
            return null;
        }

        $titleText = $title->getText();
        $baseTitle = preg_match('/^(.+?)(?: )?\([^()]+\)$/u', $titleText, $matches)
            ? trim($matches[1])
            : trim($titleText);
        if ($baseTitle === '') {
            return null;
        }

        $disambigTitle = MediaWikiServices::getInstance()
            ->getTitleFactory()
            ->makeTitle(NS_DISAMBIG, $baseTitle);

        return [
            'baseTitle' => $baseTitle,
            'exists' => $disambigTitle->exists(),
            'sourceTitle' => $title->getPrefixedText(),
            'targetTitle' => $disambigTitle->getPrefixedText(),
        ];
    }
}
