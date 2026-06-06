<?php

namespace ZetaExtension\AIEdit;

use MediaWiki\MediaWikiServices;
use User;

final class PreferencesHooks
{
    private const PREFERENCE_KEY = 'ai-edit-as-mine';

    public static function onGetPreferences(User $user, array &$preferences): bool
    {
        $userOptionsLookup = MediaWikiServices::getInstance()->getUserOptionsLookup();
        $preferences[self::PREFERENCE_KEY] = [
            'type' => 'toggle',
            'label-message' => 'ai-edit-as-mine-label',
            'default' => (bool) $userOptionsLookup->getOption($user, self::PREFERENCE_KEY, false),
            'section' => 'editing/ai-edit-as-mine',
        ];

        return true;
    }
}
