package fp

import (
	"strings"
)

// Windows limits:
//  https://learn.microsoft.com/en-us/windows/win32/fileio/naming-a-file
//	https://learn.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation?tabs=registry
// Use any character in the current code page for a name, including Unicode characters and characters in the extended character set (128–255), except for the following:
//The following reserved characters:
//   case insensitive
//   chars 0-31
//   '<', '>', ':', '"', '/', '\', '|', '?', '*'
//   Can't end with ' ' or '.'
//   Do not end a file or directory name with a space or a period. Although the underlying file system may support such names, the Windows shell and user interface does not. However, it is acceptable to specify a period as the first character of a name. For example, ".temp".
//   Do not use the following reserved names for the name of a file:
//   CON, PRN, AUX, NUL, COM1, COM2, COM3, COM4, COM5, COM6, COM7, COM8, COM9, COM¹, COM², COM³, LPT1, LPT2, LPT3, LPT4, LPT5, LPT6, LPT7, LPT8, LPT9, LPT¹, LPT², and LPT³. Also avoid these names followed immediately by an extension; for example, NUL.txt and NUL.tar.gz are both equivalent to NUL. For more information, see Namespaces.
//   To specify an extended-length path, use the "\\?\" prefix. For example, "\\?\D:\very long path".
//   The maximum path of 32,767 characters is approximate, because the "\\?\" prefix may be expanded to a longer string by the system at run time, and this expansion applies to the total length.
// Also: CLOCK$ and CONFIG$

// MacOS - APFS limits:
//  https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/APFS_Guide/FAQ/FAQ.html
// Avoid using ':' as it appears to users as '/'
// APFS accepts only valid UTF-8 encoded filenames for creation, and preserves both case and normalization of the filename on disk in all variants. APFS, like HFS+, is case-sensitive on iOS and is available in case-sensitive and case-insensitive variants on macOS, with case-insensitive being the default.
// Also normalization sensitive.
// Seems to be limited to 255 bytes in utf-8

// Btrfs limits:
// Any byte except NUL and '/'

// FreeBSD ZFS/UFS limits:
// ?

// FilenameEncodeString will encode any string to be a valid filename.
func FilenameEncodeString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "+", "_"), ":", "-")
}

// FilenameDecodeString decodes filenames encoded by FilenameEncodeString.
func FilenameDecodeString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "+", "_"), ":", "-")
}
