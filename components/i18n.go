package components

import "MusicLeCLI/state"

// headerTab returns the localized header tab label for the given nav id.
// Duplicates main.Tr for the components package (main cannot be imported).
func headerTab(id string) string {
	lang := state.Current.Language
	switch id {
	case "home":
		switch lang {
		case state.LangTurkish:
			return " Ana Sayfa "
		case state.LangSpanish:
			return " Inicio "
		case state.LangGerman:
			return " Startseite "
		case state.LangFrench:
			return " Accueil "
		case state.LangArabic:
			return " الرئيسية "
		case state.LangPortuguese:
			return " Início "
		case state.LangChinese:
			return " 首页 "
		case state.LangJapanese:
			return " ホーム "
		case state.LangItalian:
			return " Home "
		case state.LangRussian:
			return " Главная "
		default:
			return " Home "
		}
	case "downloads":
		switch lang {
		case state.LangTurkish:
			return " İndirilenler "
		case state.LangSpanish:
			return " Descargas "
		case state.LangGerman:
			return " Downloads "
		case state.LangFrench:
			return " Téléchargements "
		case state.LangArabic:
			return " التنزيلات "
		case state.LangPortuguese:
			return " Downloads "
		case state.LangChinese:
			return " 下载 "
		case state.LangJapanese:
			return " ダウンロード "
		case state.LangItalian:
			return " Download "
		case state.LangRussian:
			return " Загрузки "
		default:
			return " Downloads "
		}
	case "profile":
		switch lang {
		case state.LangTurkish:
			return " Profil "
		case state.LangSpanish:
			return " Perfil "
		case state.LangGerman:
			return " Profil "
		case state.LangFrench:
			return " Profil "
		case state.LangArabic:
			return " الملف الشخصي "
		case state.LangPortuguese:
			return " Perfil "
		case state.LangChinese:
			return " 个人资料 "
		case state.LangJapanese:
			return " プロフィール "
		case state.LangItalian:
			return " Profilo "
		case state.LangRussian:
			return " Профиль "
		default:
			return " Profile "
		}
	case "playlist":
		switch lang {
		case state.LangTurkish:
			return " Çalma Listesi "
		case state.LangSpanish:
			return " Lista "
		case state.LangGerman:
			return " Playlist "
		case state.LangFrench:
			return " Playlist "
		case state.LangArabic:
			return " قائمة التشغيل "
		case state.LangPortuguese:
			return " Playlist "
		case state.LangChinese:
			return " 播放列表 "
		case state.LangJapanese:
			return " プレイリスト "
		case state.LangItalian:
			return " Playlist "
		case state.LangRussian:
			return " Плейлист "
		default:
			return " Playlist "
		}
	case "settings":
		switch lang {
		case state.LangTurkish:
			return " Genel "
		case state.LangSpanish:
			return " General "
		case state.LangGerman:
			return " Allgemein "
		case state.LangFrench:
			return " Général "
		case state.LangArabic:
			return " عام "
		case state.LangPortuguese:
			return " Geral "
		case state.LangChinese:
			return " 通用 "
		case state.LangJapanese:
			return " 一般 "
		case state.LangItalian:
			return " Generali "
		case state.LangRussian:
			return " Общие "
		default:
			return " General "
		}
	}
	return id
}

// noTrackText returns the localized "No track playing" string.
func noTrackText() string {
	switch state.Current.Language {
	case state.LangTurkish:
		return "Çalan parça yok"
	case state.LangSpanish:
		return "Ninguna canción en reproducción"
	case state.LangGerman:
		return "Kein Titel läuft"
	case state.LangFrench:
		return "Aucun morceau en lecture"
	case state.LangArabic:
		return "لا يوجد مقطع قيد التشغيل"
	case state.LangPortuguese:
		return "Nenhuma faixa tocando"
	case state.LangChinese:
		return "当前没有播放"
	case state.LangJapanese:
		return "再生中の曲なし"
	case state.LangItalian:
		return "Nessun brano in riproduzione"
	case state.LangRussian:
		return "Ничего не играет"
	default:
		return "No track playing"
	}
}

// langBadge returns the uppercase language code (EN, TR, ES, DE, FR, AR,
// PT, ZH, JA, IT, RU) for the header status pill.
func langBadge() string {
	switch state.Current.Language {
	case state.LangTurkish:
		return "TR"
	case state.LangSpanish:
		return "ES"
	case state.LangGerman:
		return "DE"
	case state.LangFrench:
		return "FR"
	case state.LangArabic:
		return "AR"
	case state.LangPortuguese:
		return "PT"
	case state.LangChinese:
		return "ZH"
	case state.LangJapanese:
		return "JA"
	case state.LangItalian:
		return "IT"
	case state.LangRussian:
		return "RU"
	default:
		return "EN"
	}
}
