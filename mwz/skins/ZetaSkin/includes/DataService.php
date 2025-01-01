<?php
namespace ZetaSkin;

use MediaWiki\MediaWikiServices;
use ObjectCache;

class DataService {
	public static function getUserAvatar( int $userID ) {
		if ( $userID == 0 ) {
			return null;
		}
		$key = "userAvatar:$userID";
		$cached = ObjectCache::getLocalClusterInstance()->get( $key );
		if ( $cached ) {
			return $cached;
		}
		$dbr = MediaWikiServices::getInstance()->getDBLoadBalancer()->getConnection( DB_REPLICA );
		$row = $dbr->newSelectQueryBuilder()
			->select( [ 'user_name', 't', 'ghash' ] )
			->from( 'user', 'A' )
			->leftJoin( 'ldb.profiles', 'B', 'A.user_id=B.user_id' )
			->where( [ 'A.user_id' => $userID ] )
			->fetchRow();
		if ( !$row ) {
			return null;
		}
		$userAvatar = [ 'id' => $userID, 'name' => $row->user_name, 't' => $row->t, 'ghash' => $row->ghash ];
		ObjectCache::getLocalClusterInstance()->set( $key, $userAvatar );
		return $userAvatar;
	}

	public static function getContributors( $title ) {
		$data = json_decode( file_get_contents( "http://localhost/w/api.php?format=json&action=query&prop=contributors&titles=$title" ), true );
		$pages = $data['query']['pages'] ?? [];
		if ( count( $pages ) == 0 ) {
			return [];
		}
		return array_map( fn ( $x ) => self::getUserAvatar( $x['userid'] ), array_pop( $pages )['contributors'] ?? [] );
	}

	public static function getBinders( int $pageid ) {
		$url = "http://localhost/w/rest.php/binder/$pageid";
		$contents = file_get_contents( $url );
		$binders = json_decode( $contents, true );
		return $binders;
	}
}
