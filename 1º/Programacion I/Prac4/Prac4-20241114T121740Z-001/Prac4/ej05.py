def leer_lista_enteros():
    lista = []
    cadena = input('Introduce un número: ')
    while cadena != ' ':
        numero = int(cadena)
        lista.append(numero)
        cadena = input('Introduce otro número: ')
    return lista

def borrar_elementos(lista, posicion):
        if posicion > len(lista) or posicion < 0:
            return False
        else:
            del lista[posicion]
            return True


lista = leer_lista_enteros()
print(lista)

while lista:
    posicion = int(input('Introduce la posición que deseas eliminar: '))
    if borrar_elementos(lista, posicion):
        print(f'Elemento de la lista borrado. Esta es la lista actualizada {lista}')

print('La lista está vacia')


