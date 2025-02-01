<?php

namespace ZetaExtension\Map;

class Map
{
    public static function onParserFirstCallInit($parser)
    {
        $parser->setHook('map', [self::class, 'render']);
    }

    public static function render($text, array $args, $parser, $frame)
    {
        $type = 1;
        $zoom = 5;
        $width = 720;
        $height = 360;
        $place = $parser->recursiveTagParse($text, $frame);
        if (array_key_exists('type', $args)) {
            $type = $parser->recursiveTagParse($args['type'], $frame);
        }
        if (array_key_exists('zoom', $args)) {
            $zoom = $parser->recursiveTagParse($args['zoom'], $frame);
        }
        if (array_key_exists('width', $args)) {
            $width = $parser->recursiveTagParse($args['width'], $frame);
        }
        if (array_key_exists('height', $args)) {
            $height = $parser->recursiveTagParse($args['height'], $frame);
        }
        $parser->getOutput()->addModules("map{$type}");

        return "<div class='zmap{$type}' place='$place' zoom='$zoom' style='width:{$width}px;height:{$height}px'></div>";
    }
}
