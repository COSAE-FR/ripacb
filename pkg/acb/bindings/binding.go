package bindings

import (
	"github.com/COSAE-FR/ripacb/pkg/acb/config"
	"mime/multipart"
)

/*
"userkey=" . $userkey .
		"&revision=" . urlencode($_REQUEST['newver']) .
		"&version=" . $g['product_version'] .
		"&uid=" . urlencode($uniqueID));
*/
type GetBackupRequest struct {
	Uid       string `form:"uid"`
	Version   string `form:"version"`
	Revision  string `form:"revision"`
	DeviceKey string `form:"userkey"`
}

type StatusResponse struct {
	Code     int              `json:"code"`
	Message  string           `json:"message"`
	Features *config.Features `json:"features,omitempty"`
}

//
// $post_fields = array(
//				'reason' => htmlspecialchars($reason),
//				'uid' => $uniqueID,
//				'file' => curl_file_create($tmpname, 'image/jpg', 'config.jpg'),
//				'userkey' => htmlspecialchars($userkey),
//				'sha256_hash' => $raw_config_sha256_hash,
//				'version' => $g['product_version'],
//				'hint' => $config['system']['acb']['hint'],
//				'manmax' => $manmax
//			);
type SaveBackupRequest struct {
	Uid       string                `form:"uid"`
	Version   string                `form:"version"`
	DeviceKey string                `form:"userkey"`
	Reason    string                `form:"reason"`
	Content   *multipart.FileHeader `form:"file"`
	Hash      string                `form:"sha256_hash"`
	Hint      string                `form:"hint"`
	ManualMax int                   `form:"manmax"`
}
