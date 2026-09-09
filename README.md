# RamDB (sistema de armazenamento de dados em memória primária)

## Protocolo de Comunicação com o Servidor:

### Anatomia da de uma mensagem:
O servidor da DB escuta em uma porta TCP. As mensagens enviadas para o servidor devem apresentar os seguinte formato:

- bytes [0:4]:
  Informam do tamanho da mensagem, contando com os primeiros quatro bytes. Esses bytes são decodificados no formato little endian como um uint32.
- bytes [4:]:
  O restante dos bytes constituem o corpo/conteúdo da mensagem propriamente dita. Esses bytes são decodificados no formato UTF-8
