<?php

namespace ZetaSkin;

use eftec\bladeone\BladeOne;
use Html;
use SkinTemplate;
use Title;

class SkinBlade extends SkinTemplate
{
    private $templateParser = null;

    public function generateHTML()
    {
        $this->setupTemplateContext();
        $out = $this->getOutput();
        $blade = new BladeOne(__DIR__.'/views', __DIR__.'/cache', BladeOne::MODE_AUTO);
        $html = $out->headElement($this);
        $html .= $blade->run('app', $this->getTemplateData());
        $html .= $out->tailElement($this);

        return $html;
    }

    public function getTemplateData()
    {
        $data = parent::getTemplateData();
        $data = array_combine(array_map(fn ($x) => str_replace('-', '_', $x), array_keys($data)), $data);
        $out = $this->getOutput();
        $printSource = Html::rawElement('div', ['class' => 'printfooter', 'data-nosnippet' => ''], $this->printSource());
        $bodyContent = $out->getHTML()."\n".$printSource;
        $newTalksHtml = $this->getNewtalks() ?: null;
        $data += [
            'array_indicators' => $this->getIndicatorsData($out->getIndicators()),
            'html_site_notice' => $this->getSiteNotice() ?: null,
            'html_user_message' => $newTalksHtml ? Html::rawElement('div', ['class' => 'usermessage'], $newTalksHtml) : null,
            'html_subtitle' => $this->prepareSubtitle(),
            'html_body_content' => $this->wrapHTML($out->getTitle(), $bodyContent),
            'html_categories' => $this->getCategories(),
            'html_after_content' => $this->afterContentHook(),
            'html_undelete_link' => $this->prepareUndeleteLink(),
            'html_user_language_attributes' => $this->prepareUserLanguageAttributes(),
            'link_mainpage' => Title::newMainPage()->getLocalURL(),
        ];
        foreach ($this->options['messages'] ?? [] as $message) {
            $data["msg_{$message}"] = $this->msg($message)->text();
        }

        return $data;
    }
}
