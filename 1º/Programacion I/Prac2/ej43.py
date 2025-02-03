num =  int(input("Dime un numero:   "))

cont = 0
numero = 1
divisores_cantidad = 0
numero_divisores = 0

while cont < num:
    resp = 0

    for j in range(1, numero+1):
        if numero % j == 0:
            resp = resp + 1

    if resp  >= divisores_cantidad:
        numero_divisores = numero
        divisores_cantidad =  resp

    cont+=1
    numero+=1


print(f"El numero con mas divisores es {numero_divisores} ({divisores_cantidad} divisores)")

