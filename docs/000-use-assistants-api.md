# Title
Assistants APIの導入

## Summary
Botの使用するAPIをAssistantsAPIに変更する作業手順や実装方針をまとめる

## Background
<!-- このDesinDocの背景を説明(参考となるリンクなどを貼るだけでも良い) -->
AssistantsAPIがリリースされたので、これを使用したい
https://platform.openai.com/docs/assistants/overview

## Detail Design
<!-- このDesinDocの設計内容を説明
基本的な方針やクラス図、使うAPIやデザインパターン、データフローなど -->
### 変更点
- ステージング機能
    - 現状、mainマージすると本番環境に影響が出るが、これを特定のチャンネル(genAI/sandbox)のみ、別のコードが機能するようにすることができるようにする

- threads
    - 現状Userごとの会話はチャンネル単位でDBによって管理されているが、
これをAssistantsAPIが提供するthreadsに置き換える。

- Code Interpreter
    - コードを向こうで実行して結果を得られる。
    - 結果はテキストのほか画像やデータファイル(csvなど)としてリンクが返ってくる

### 実装方針
- ステージング
    - handlerで振り分ける
- threads
    - DBにチャンネルIDとthreadのペアをもつor連想配列でもつ
    - 連想配列でもいい気がする(Bot再起動したら全部リセットでも困らなさそう)
- Code Interpreter
    - こっちで実装すべきは注釈の対応
    - messageとは別にannotationの配列が来るらしいので、それを適切に文字列に反映する必要がある
    - https://platform.openai.com/docs/assistants/deep-dive/managing-threads-and-messages

## Alternatives
<!-- 他の選択肢があれば説明 -->
以下二つの新機能は一旦最初のAssistantsAPIの導入時にはいれない
- file-search
- function-calling

理由としては、ファイルや関数の管理をどうするかなど実装以外の検討が進んでいないため


## Q&A
<!-- 質問と回答(レビュー時に出たものもここに追記する) -->