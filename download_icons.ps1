$baseUrl = "https://img.icons8.com/ios-filled/50/"
$blue = "0b2b5a/"
$green = "4CAF50/"
$red = "F44336/"

$icons = @(
    @("user.png", $blue, "user.png"),
    @("passport.png", $blue, "passport.png"),
    @("briefcase.png", $green, "briefcase.png"),
    @("globe.png", $blue, "globe.png"),
    @("chat.png", $blue, "chat.png"),
    @("checkmark.png", $green, "check.png"),
    @("multiply.png", $red, "cross.png"),
    @("baby.png", $green, "baby.png"),
    @("broom.png", $green, "broom.png"),
    @("iron.png", $green, "iron.png"),
    @("cooking-pot.png", $green, "pot.png"),
    @("open-book.png", $green, "book.png"),
    @("holding-heart.png", $green, "care.png"),
    @("flag.png", $blue, "flag.png"),
    @("church.png", $blue, "church.png"),
    @("calendar.png", $blue, "calendar.png"),
    @("marker.png", $blue, "marker.png"),
    @("female.png", $blue, "female.png"),
    @("hearts.png", $blue, "hearts.png"),
    @("children.png", $blue, "children.png"),
    @("graduation-cap.png", $blue, "education.png")
)

foreach ($icon in $icons) {
    $url = $baseUrl + $icon[1] + $icon[0]
    $dest = "c:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\internal\assets\" + $icon[2]
    Invoke-WebRequest -Uri $url -OutFile $dest
}
