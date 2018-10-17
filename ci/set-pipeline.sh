#!/bin/bash

if [[ "$(uname)" == "Darwin" ]]; then
  readlink="greadlink"
else
  readlink="readlink"
fi

script_dir="$($readlink -f "$(dirname "$0")")"

if ! lpass status; then
  echo "You need to log in to LastPass first. Try lpass login [email]."
  exit 1
fi

username="$(lpass show "${LASTPASS_NOTE_ID:-2874932944274242926}" --username)"
password="$(lpass show "${LASTPASS_NOTE_ID:-2874932944274242926}" --password)"

varfile="$(mktemp)"
trap 'rm ${varfile}' 0 1 2 3 6 15

cat <<EOF > "$varfile"
bosh-stemcells-cf:
  username: ${username}
  password: ${password}
EOF

fly -t "${FLY_TARGET:-"wings"}" sp -p "${PIPELINE_NAME:-"bosh-stemcells"}" -c "${script_dir}/pipeline.yml" -l "$varfile"
