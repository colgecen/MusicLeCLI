# MusicLe CLI

Terminalde çalışan, Spotify esintili arayüzlü müzik çalar ve kişisel arşiv yöneticisi (Go + Bubble Tea).

![MusicLe CLI](assets/MusicleCLI-app.png)

## Özellikler

- Spotify ve YouTube bağlantısıyla parça ve çalma listesi indirme (gömülü yt-dlp + FFmpeg, MP3 320k + kapak resmi)
- Şarkı listesi, profil ve çalma listesi yönetimi; kayıtlar `song_list.txt` içinde tutulur
- Çalan şarkının kapağı ana ekranda otomatik gösterilir
- Spektrum görselleştirme, kayan şarkı sözü ve ses seviyesi çubukları
- 11 dil desteği: Türkçe, English, Español, Deutsch, Français, العربية, Português, 中文, 日本語, Italiano, Русский
- Tema ve spektrum paleti seçenekleri, çıkış cihazı ve ses limiti ayarı
- Tarayıcı bağlayıcı: açık Spotify/YouTube Music sekmesindeki listeleri içe aktarma
- Klavye kısayolları: F1 odak değiştir, F2 görünüm değiştir, F3 ayar sekmesi, ↑↓ ses, ←→ 5sn sar, Boşluk çal/duraklat, Esc ana sayfa
- Tek dosyalık kurulum: AppImage / .exe / tar.gz, Windows/macOS/Linux desteği

## Kurulum

Projeyi yerel ortamınızda çalıştırmak için aşağıdaki adımları takip edin:

```bash
git clone https://github.com/colgecen/MusicLeCLI.git
cd MusicLeCLI
go build -o musicle-cli .
```

Menülü derleme aracıyla paket çıkarmak için:

```bash
musiclecli
```

Hedef ve format seçin (Linux: AppImage/RPM/tar.gz/binary, Windows: .exe/.exe+zip, macOS: tar.gz/binary).
Hazır ikililer için [sürümler](https://github.com/colgecen/MusicLeCLI/releases) sayfasına bakabilirsiniz.

## Kullanım

Projenin nasıl kullanılacağına dair örnek:

```bash
./build/musicle-cli
```

İndirme sekmesine Spotify veya YouTube bağlantısını yapıştırıp hedef listeyi seçin, indirme bitince şarkılar kütüphanede belirir. Ana ekranda `↑↓` ile listede gezinip `Enter` ile çalın, `F1` ile oynatma çubuğuna odaklanıp ses ve sarma yapın. Ayarlardan dil, tema ve ses çıkışını değiştirin.

Yapılandırma `~/.config/musicle/config.json`, müzik arşivi `~/Music/MusicLe/` altında tutulur.

## Lisans

Bu proje [MIT](LICENSE) lisansı altında lisanslanmıştır.
