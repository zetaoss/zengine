$.getScript('//openapi.map.naver.com/openapi/v3/maps.js?ncpClientId=tu2p9fjuwb', () => {
  $('.zmap2').each(function (i) {
    const vm = $(this);
    const target = 'zmap2-' + i;
    const zoom = 1 * vm.attr('zoom');
    const place = vm.attr('place');
    vm.prop('id', target)
    $.get('/api/geo2/' + place).done((r) => {
      const center = new naver.maps.LatLng(r[1], r[0]);
      var map = new naver.maps.Map(target, { center: center, zoom: zoom, zoomControl: true });
      new naver.maps.Marker({ position: center, map: map });
    })
  })
})
