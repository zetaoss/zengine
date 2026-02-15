#!/bin/bash
cd $(dirname $0)

npm install

PASTEL_VERSION=0.9.0
which pastel || (curl -LO "https://github.com/sharkdp/pastel/releases/download/v${PASTEL_VERSION}/pastel_${PASTEL_VERSION}_amd64.deb"; apt install ./pastel_${PASTEL_VERSION}_amd64.deb; rm -f pastel_${PASTEL_VERSION}_amd64.deb)
which prettier || npm install -g prettier

echo
echo ==== 0. init
echo
modulesArray=(
  mediawiki.rcfilters.filters.ui
  mediawiki.special.preferences.styles.ooui
  mediawiki.special.search.styles
)

modules=""
for module in "${modulesArray[@]}"; do
	modules="$modules%7C$module"
done
modules=$(echo $modules | cut -c 4-)
curl -sL "http://zeta/w/load.php?lang=ko&only=styles&modules=$modules" -o modules.scss

prettier modules.scss --write --print-width 9999

sed ':a;N;$!ba;s|,\n\s*|, |g' -i modules.scss
sed 's|rgba|#rgba|g'          -i modules.scss
sed 's|^  [^#]\+;$||g'        -i modules.scss
sed 's|#rgba|rgba|g'          -i modules.scss

sed ':a;N;$!ba;s|{\s*}|{}|g'  -i modules.scss
sed 's|^.* {}$||g'            -i modules.scss
sed ':a;N;$!ba;s|{\s*}|{}|g'  -i modules.scss
sed 's|^.* {}$||g'            -i modules.scss
prettier modules.scss --write --print-width 9999

echo
echo ==== 1. filelist
echo
mkdir -p tmp
ls /app/w/resources/lib/ooui/oojs-ui-*-wikimediaui.css modules.scss > tmp/1_filelist.txt

echo
echo ==== 2. colors
echo
cat /dev/null > tmp/2_colors.txt.tmp
while read FILE; do
  echo $FILE
  BAK=$FILE.bak
  if [ -f $BAK ]; then
    cp -a $BAK $FILE
  else
    cp -a $FILE $BAK
  fi
  cat $BAK | sed 's|[ \n;]|\n|g' | grep -o '^#[0-9a-f]\+'             >> tmp/2_colors.txt.tmp
  cat $BAK | sed 's|, |,|g' | grep -o 'rgba([^)]\+)' | sed 's|,|, |g' >> tmp/2_colors.txt.tmp
done < tmp/1_filelist.txt
cat   tmp/2_colors.txt.tmp | sort | uniq > tmp/2_colors.txt
rm -f tmp/2_colors.txt.tmp
rm -f modules.scss.bak

##### functions
function proc_hex {
  H=$(pastel color "$INPUT" | pastel format hsl-hue)
  S=$(pastel color "$INPUT" | pastel format hsl-saturation)
  S=$(echo $S | awk '{printf "%.1f",$1*100}')
  L=$(pastel color "$INPUT" | pastel format hsl-lightness)
  L=$(echo $L | awk '{printf "%.1f",100-$1*100}')
  NAME=$(echo $INPUT | sed 's|#||g')
  LIGHT=$INPUT
  DARK=$(pastel color "hsl($H, $S%, $L%)" | pastel format hex)
}

function proc_rgba {
  H=$(pastel color "$INPUT" | pastel format hsl-hue)
  S=$(pastel color "$INPUT" | pastel format hsl-saturation)
  S=$(echo $S | awk '{printf "%.1f",$1*100}')
  L=$(pastel color "$INPUT" | pastel format hsl-lightness)
  L=$(echo $L | awk '{printf "%.1f",100-$1*100}')
  A=$(echo "$INPUT" | grep -o ', [0-9.]\+)' | grep -o '[0-9.]\+')
  NAME=$(pastel color "$INPUT" | pastel format hex | sed 's|#||g')
  LIGHT=$(pastel color "$INPUT" | pastel format hex)
  DARK=$(pastel color "hsla($H, $S%, $L%, $A)" | pastel format hex)
}

echo
echo ==== 3. table
echo
cat /dev/null > tmp/3_table.tsv
while read INPUT; do
  if [[ $INPUT == \#* ]]; then
    proc_hex
  else
    proc_rgba
  fi
  echo -e "$NAME\t$LIGHT\t$DARK\t$INPUT" >> tmp/3_table.tsv
done < tmp/2_colors.txt

echo
echo ==== 4. variables
echo
echo ':root {'                     >  variables.scss
while read -r NAME LIGHT DARK INPUT; do
  echo "  --uc-$NAME: $LIGHT;"     >> variables.scss
done < tmp/3_table.tsv
echo '}'                           >> variables.scss
echo                               >> variables.scss
echo '.dark {'                     >> variables.scss
while read -r NAME LIGHT DARK INPUT; do
  echo "  --uc-$NAME: $DARK;"      >> variables.scss
done < tmp/3_table.tsv
echo '}'                           >> variables.scss

echo
echo ==== 5. replace files
echo
while read FILE; do
  while read -r NAME LIGHT DARK INPUT; do
    sed "s|$INPUT;|var(--uc-$NAME);|g" -i $FILE
  done < tmp/3_table.tsv
done < tmp/1_filelist.txt
