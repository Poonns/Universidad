def leer_lista_enteros():
    lista = []
    cadena = input('Introduce un número: ')
    while cadena != ' ':
        numero = int(cadena)
        lista.append(numero)
        cadena = input('Introduce otro número: ')
    return lista

def borrar_elementos(lista, elemento):
        for posicion in range(len(lista)-1, -1, -1):
            if elemento == lista[posicion]:
                del lista[posicion]

                return True

        return False

# otra posible solución:
# posicion = 0
# while posicion < len(lista):
#     if elemento == lista[posicion]
#         del lista[posicion]
#     else:
#         posicion += 1


lista = leer_lista_enteros()
print(lista)

while lista!=[]:
    elemento = int(input('Introduce el elemento que deseas eliminar: '))
    resultado = borrar_elementos(lista, elemento)
    if resultado == True:
        print(f'Elemento de la lista borrado. Esta es la lista actualizada {lista}')
    else:
        print('El elemento no se encuentra en la lista')

print('La lista está vacia')