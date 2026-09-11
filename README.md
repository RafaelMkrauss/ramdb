# RamDB (sistema de armazenamento de dados em memória primária)

## Protocolo de Comunicação com o Servidor:

### Anatomia da de uma mensagem:
O servidor da DB escuta em uma porta TCP. As mensagens enviadas para o servidor devem apresentar os seguinte formato:

- Padding: bytes [0:4] ~ [0x1, 0x1, 0x1, 0x1]:
  Os primeiros quatro bytes da mensagem consistem em uma sequência de quatro bytes com valor 0x1. Servem para identificar o começo da mensagem.
- Message Size: bytes [4:8]:
  Informam do tamanho do corpo da mensagem (somente o corpo). Esses bytes são decodificados no formato little endian como um uint32.
- Message Id: bytes [8:12]:
  Quatro bytes que servem como identificador identificador único da requisição. Na resposta do servidor ao cliente esse mesmo valor é enviado na header.
- Body: bytes [12:]:
  O restante dos bytes constituem o corpo/conteúdo da mensagem propriamente dita. Esses bytes são decodificados no formato UTF-8
