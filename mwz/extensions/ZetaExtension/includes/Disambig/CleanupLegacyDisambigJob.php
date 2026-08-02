<?php

namespace ZetaExtension\Disambig;

use Job;
use MediaWiki\CommentStore\CommentStoreComment;
use MediaWiki\Content\ContentHandler;
use MediaWiki\Content\TextContent;
use MediaWiki\Page\WikiPageFactory;
use MediaWiki\Revision\SlotRecord;
use MediaWiki\Title\Title;
use MediaWiki\User\User;
use MediaWiki\User\UserFactory;

final class CleanupLegacyDisambigJob extends Job
{
    public const TYPE = 'zetaCleanupLegacyDisambig';

    private UserFactory $userFactory;
    private WikiPageFactory $wikiPageFactory;

    public function __construct(
        Title $title,
        array $params,
        UserFactory $userFactory,
        WikiPageFactory $wikiPageFactory
    ) {
        parent::__construct(self::TYPE, $title, $params);
        $this->userFactory = $userFactory;
        $this->wikiPageFactory = $wikiPageFactory;
    }

    public function run(): bool
    {
        $wikiPage = $this->wikiPageFactory->newFromTitle($this->title);
        $redirectTarget = trim((string) ($this->params['redirect_target'] ?? ''));
        if ($redirectTarget !== '') {
            $exists = $wikiPage->exists();
            if ($exists && ! $wikiPage->isRedirect()) {
                return true;
            }

            $user = $this->resolveUser();
            if ($user === null) {
                return false;
            }

            $updater = $wikiPage->newPageUpdater($user);
            if ($exists) {
                $revision = $updater->grabParentRevision();
                if ($revision === null) {
                    return true;
                }
                $content = $revision->getContent(SlotRecord::MAIN);
                $currentTarget = $content !== null ? $content->getRedirectTarget() : null;
                if ($currentTarget !== null && $currentTarget->getPrefixedText() === $redirectTarget) {
                    return true;
                }
            }
            $updater->setContent(
                SlotRecord::MAIN,
                ContentHandler::makeContent("#REDIRECT [[{$redirectTarget}]]\n", $this->title, CONTENT_MODEL_WIKITEXT)
            );
            $disambigTitle = (string) ($this->params['disambig_title'] ?? 'Disambig 문서');
            $updater->saveRevision(
                CommentStoreComment::newUnsavedComment("{$disambigTitle}의 첫 항목으로 넘겨주기 ".($exists ? '갱신' : '생성')),
                EDIT_MINOR
            );

            if (! $updater->wasSuccessful()) {
                $this->setLastError('동음이의 대표 제목의 넘겨주기 처리에 실패했습니다.');
                return false;
            }

            return true;
        }

        if (! $wikiPage->exists() || $wikiPage->isRedirect()) {
            return true;
        }

        $user = $this->resolveUser();
        if ($user === null) {
            return false;
        }

        $updater = $wikiPage->newPageUpdater($user);
        $revision = $updater->grabParentRevision();
        if ($revision === null) {
            return true;
        }

        $content = $revision->getContent(SlotRecord::MAIN);
        if (! $content instanceof TextContent) {
            return true;
        }

        $current = $content->getText();
        $cleaned = self::stripLegacyDisambigTemplates($current);
        if ($cleaned === $current) {
            return true;
        }

        $updater->setContent(
            SlotRecord::MAIN,
            ContentHandler::makeContent($cleaned, $this->title, $content->getModel())
        );
        $disambigTitle = (string) ($this->params['disambig_title'] ?? 'Disambig 문서');
        $updater->saveRevision(
            CommentStoreComment::newUnsavedComment("{$disambigTitle}로 동음이의 관계 통합"),
            EDIT_MINOR
        );

        if (! $updater->wasSuccessful()) {
            $this->setLastError('다른뜻 틀 제거 편집에 실패했습니다.');
            return false;
        }

        return true;
    }

    private function resolveUser(): ?User
    {
        $userId = (int) ($this->params['user_id'] ?? 0);
        $userName = (string) ($this->params['user_name'] ?? '');
        $user = $userId > 0
            ? $this->userFactory->newFromId($userId)
            : $this->userFactory->newFromName($userName);
        if ($user === null || $user->getName() === '') {
            $this->setLastError('동음이의 문서 정리 작업의 편집 사용자를 찾을 수 없습니다.');
            return null;
        }

        return $user;
    }

    public static function stripLegacyDisambigTemplates(string $wikitext): string
    {
        return (string) preg_replace(
            '/^[\t ]*\{\{\s*다른[\s_]*뜻\s*(?:\|[^{}]*?)?\}\}[\t ]*(?:\r?\n|$)/imu',
            '',
            $wikitext
        );
    }
}
