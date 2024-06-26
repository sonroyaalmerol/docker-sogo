#!/bin/bash

customConfigFolder="/etc/sogo/sogo.conf.d/"
configPath="/etc/sogo/sogo.conf"
yamlPath="/etc/sogo/sogo.yaml"

Print() {
  local indent=$(($1 * 2))
  local text=$2
  printf "%*s%s" $indent '' "$text" >> "$configPath"
}

Println() {
  local indent=$(($1 * 2))
  local text=$2
  printf "%*s%s\n" $indent '' "$text" >> "$configPath"
}

ParsePrimitives() {
    local value="$1"

    if [ "$value" = "true" ]; then
      echo -n "YES"
    elif [ "$value" = "false" ]; then
      echo -n "NO"
    elif echo "$value" | grep -qE '^[0-9]+$'; then
      echo -n "$value"
    else
      echo -n "\"$value\""
    fi
}

ParseValues() {
  local key="$1"
  local path="$2"
  local level="$3"

  local value
  value=$(yq "$path" "$yamlPath")
  local valueType
  valueType=$(yq "$path | type" "$yamlPath" | xargs)

  if [ "$valueType" = '!!seq' ]; then
    Println $((level + 1)) "$key = ("
    local i=0
    while true; do
      local item
      item=$(yq "$path | .[$i]" "$yamlPath")
      
      if [ "$item" = "null" ]; then
        break
      fi

      if [ $i -ne 0 ]; then
        Println 0 ","
      fi

      local itemType
      itemType=$(yq "$path | .[$i] | type" "$yamlPath" | xargs)

      if [ "$itemType" != '!!seq' ] && [ "$itemType" != '!!map' ]; then
        Print $((level + 2)) "$(ParsePrimitives "$item")"
      else
        ParseValues "$i" "$path | .[$i]" $((level + 1))
      fi
      i=$((i+1))
    done
    
    Println 0 ""
    Println $((level + 1)) ");"
  elif [ "$valueType" = '!!map' ]; then
    if echo "$key" | grep -qE '^[0-9]+$'; then
      Println $((level + 1)) "{"
    else
      Println $((level + 1)) "$key = {"
    fi
    
    yq "$path | keys | .[]" "$yamlPath" | while IFS= read -r subKey; do
      ParseValues "$subKey" "$path | .$subKey" $((level + 1))
    done
    Print $((level + 1)) "}"
  elif [ "$valueType" = '!!str' ] || [ "$valueType" = '!!int' ] || [ "$valueType" = '!!bool' ]; then
    Println $((level + 1)) "$key = $(ParsePrimitives "$value");"
  else
    Println $((level + 1)) "/* Invalid type for $value: $valueType */"
  fi
}

GenerateConfigFile() {
  # Ensure all .yml extensions are .yaml
  for file in "$customConfigFolder"/*.yml; do
    if [ -e "$file" ]; then
      mv "$file" "${file%.yml}.yaml"
    fi
  done

  # Reset config file
  > "$configPath"

  yq eval-all \
    '. as $item ireduce ({}; . * $item )' \
    "$customConfigFolder"*".yaml" > "$yamlPath"

  disclaimerMessage=$(cat <<EOF
  /* *********************  Main SOGo configuration file  **********************
  *                                                                           *
  * This configuration is AUTOGENERATED by the Docker container based on the  *
  * YAML files provided in /etc/sogo/sogo.conf.d/.                            *
  *                                                                           *
  * YAML configurations is only applicable for this specific container.       *
  *                                                                           *
  * The script inside the container merges the YAML files in the directory    *
  * and converts the merged config into OpenStep plist format.                *
  *                                                                           *
  * Since the content of this file is a dictionary in OpenStep plist format,  *
  * the curly braces enclosing the body of the configuration are mandatory.   *
  * See the Installation Guide for details on the format.                     *
  *                                                                           *
  * C and C++ style comments are supported.                                   *
  *                                                                           *
  * This example configuration contains only a subset of all available        *
  * configuration parameters. Please see the installation guide more details. *
  *                                                                           *
  * ~sogo/GNUstep/Defaults/.GNUstepDefaults has precedence over this file,    *
  * make sure to move it away to avoid unwanted parameter overrides.          *
  *                                                                           *
  * **************************************************************************/
EOF
)

  Println 0 "{"
  Println 0 "$disclaimerMessage"
  Println 0 ""
  yq 'keys | .[]' "$yamlPath" | sort -u | while IFS= read -r rootKey; do
    ParseValues "$rootKey" ".$rootKey" 0
  done
  Print 0 "}"
}
