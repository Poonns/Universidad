n = int(input('Introduce un numero: '))
contador = 0
suma = 0

while n >= 0:
    suma += n
    contador += 1
    n = int(input('Introduce un numero: '))

media = suma / contador
print(media)

