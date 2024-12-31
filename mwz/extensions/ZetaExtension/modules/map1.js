$.getScript('//cdnjs.cloudflare.com/ajax/libs/openlayers/4.6.5/ol.js', () => {
  $('.zmap1').each(function (i) {
    const vm = $(this);
    const target = 'zmap1-' + i;
    const zoom = vm.attr('zoom');
    const place = vm.attr('place');
    vm.prop('id', target)
    $.get('/api/geo1/' + place).done((r) => {
      const center = ol.proj.fromLonLat([1 * r[0], 1 * r[1]]);
      new ol.Map({
        target: target,
        layers: [
          new ol.layer.Tile({ source: new ol.source.OSM() }),
          new ol.layer.Vector({
            source: new ol.source.Vector({ features: [new ol.Feature({ geometry: new ol.geom.Point(center) })] }),
            style: new ol.style.Style({
              image: new ol.style.Icon({
                anchor: [0.5, 30], anchorXUnits: 'fraction', anchorYUnits: 'pixels', src: '//ssl.pstatic.net/static/maps/mantle/1x/marker-default.png'
                //src:'data:image/svg+xml,<svg width="30" height="30" viewBox="0 0 1792 1792" xmlns="http://www.w3.org/2000/svg"><path d="M1152 640q0-106-75-181t-181-75-181 75-75 181 75 181 181 75 181-75 75-181zm256 0q0 109-33 179l-364 774q-16 33-47.5 52t-67.5 19-67.5-19-46.5-52l-365-774q-33-70-33-179 0-212 150-362t362-150 362 150 150 362z"/></svg>'
              })
            })
          })],
        view: new ol.View({ center: center, zoom: zoom }),
        interactions: ol.interaction.defaults({ mouseWheelZoom: false })
      })
    })
  })
})
