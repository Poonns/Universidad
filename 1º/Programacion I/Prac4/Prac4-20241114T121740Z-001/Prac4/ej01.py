lista = []
cadena = input('Introduce una cadena: ')
while cadena != ' ':  
    lista.append(cadena)
    cadena = input('Introduce otra cadena: ')

#Mostrar cada cadena en una línea diferente
for indice in range(len(lista)):
    print(lista[indice])