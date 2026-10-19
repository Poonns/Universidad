;Sin definir otras funciones adicionales, implementa una función cuenta_exitos que reciba una función de evaluación (que devuelve T o NIL) y un número arbitrario de argumentos. Debe devolver un único número indicando cuántos elementos cumplen dicha condición.
;(cuenta_exitos #'evenp 1 2 3 4 5 6) -> 3
;Tu objetivo: Resolverlo en una sola línea usando tu combo de mapcar.
(defun cuenta_exitos(f &rest a)
    (length(remove NIL(mapcar f a)))
)

;Sin definir otras funciones adicionales, implementa una función emparejar_con_fijo que, dado un elemento fijo y un número arbitrario de argumentos adicionales, devuelva una lista de listas donde cada argumento esté emparejado con el elemento fijo en la primera posición.
;(emparejar_con_fijo 'X 1 2 3) -> ((X 1) (X 2) (X 3))
;Tu objetivo: Resolverlo de forma estrictamente recursiva (sin mapcar). Pon mucha atención a dónde usas cons y dónde usas list para no anidar de más.
(defun emparejar_con_fijo(e &rest a)
    (cond  
        ((null a)nil)
        (t
            (cons(list e (car a)(apply #'emparejar_con_fijo e (cdr a))))
        )
    )
)

;Sin definir otras funciones adicionales ni usar variables locales (let), implementa una función separar_pares_impares que reciba un número arbitrario de enteros y devuelva una lista con dos sublistas: la primera con los números pares (evenp) y la segunda con los impares (oddp).
;(separar_pares_impares 1 2 3 4 5) -> ((2 4) (1 3 5))
;(separar_pares_impares 2 4) -> ((2 4) NIL)
(defun separar_pares_impares(&rest a)
    (cond
        ((null a)(list nil nil))
        ((zerop (car a))
            (apply #'separar_pares_impares (cdr a))
        )
        ((if(=(mod 2 (car a))0))
            (list
                (cons(car a)(car(apply #'separar_pares_impares (cdr a))))
                (cadr(apply #'separar_pares_impares (cdr a)))
            )
        )
        (t
            (list
                (car(apply #'separar_pares_impares (cdr a)))
                (cons(car a)(cadr(apply #'separar_pares_impares (cdr a))))
            )
        )
    )
)
,
;Reto 1: La Matriz Cruzada (Anidamiento Funcional vs Recursividad de Doble Capa)
;Sin definir otras funciones adicionales, implementa multiplicadores_locos que reciba una lista de números "base" y un número arbitrario de multiplicadores sueltos. Debe devolver una lista de listas, donde cada número base se haya multiplicado por todos los multiplicadores.
;(multiplicadores_locos '(1 2 3) 10 20) -> ((10 20) (20 40) (30 60))
;(multiplicadores_locos '(5) 2 3 4) -> ((10 15 20))

;Version mapcar
(defun multiplicadores_locos (l &rest m)
    (mapcar(lambda(x)((list(mapcar(lambda(multiplicador)(* multiplicador x))m))))l)
)

;Version recursiva
(defun multiplicadores_locos (l &rest m)
    (cond
        ((null m)l)
        (t
            (cons(mapcar(lambda(x)(* (car l) x)))(apply #'multiplicadores_locos (cdr l) m))
        )
    )
)

;Sin definir otras funciones adicionales ni usar variables locales (let), implementa dividir_transformar que reciba un número arbitrario de enteros. Debe devolver una lista con dos sublistas: la primera contendrá los números pares multiplicados por 10, y la segunda contendrá los números impares restándoles 1 (puedes usar la primitiva 1-).
;(dividir_transformar 1 2 3 4 5) -> ((20 40) (0 2 4))

;Version mapcar
(defun dividir_transformar (&rest enteros)
    (list
        (remove nil(mapcar (lambda (x)(if(=(mod x 2)0)(* x 10) nil)) enteros))
        (remove nil (mapcar (lambda (x)(if(=(mod x 2)1)(1- x) nil)) enteros))
    )
)

;Version recursiva
(defun dividir_transforma (&rest enteros)
    (cond
        ((null enteros)(list nil nil))
        ((= (mod (car enteros) 2)0)
            (list
                (cons (* 10 (car enteros)) (car(apply #'dividir_transforma (cdr enteros))))
                                   (cadr(apply #'dividir_transforma (cdr enteros)))
            )
        )
        (t
            (list
                                        (car(apply #'dividir_transforma (cdr enteros)))
                (cons (1- (car enteros))(cadr(apply #'dividir_transforma (cdr enteros))))
            )
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada aplica_si que reciba una función de condición (predicado), una función de transformación y un número arbitrario de listas. Debe devolver una lista con el mismo número de sublistas, donde a los elementos individuales que cumplan la condición se les haya aplicado la transformación, y los que no, se queden exactamente igual.
;(aplica_si #'evenp (lambda (x) (* x 10)) '(1 2 3) '(4 5) '(6)) → ((1 20 3) (40 5) (60))
(defun aplica_si (f_cond f_trans &rest l)
    (mapcar(lambda(sublista)(mapcar(lambda(y)(if(funcall f_cond y)(funcall f_trans y) y))sublista))l)
)

;Siguiendo la lógica del ejercicio contarCondiciones de la segunda convocatoria, pero un paso más allá. Sin definir otras funciones adicionales, implementar una función en lisp llamada clasificar_valores que, dada una lista sin anidamientos de valores y un número arbitrario de funciones que representen condiciones, devuelva una lista de listas. Cada sublista debe contener los valores exactos de la lista original que cumplen con la condición correspondiente.
;(clasificar_valores '(1 2 3 4) #'evenp #'oddp (lambda (x) (> x 2))) → ((2 4) (1 3) (3 4))
(defun clasisicar_valores (l &rest f_cond)
    (cond
        ((null f_cond)(list nil))
        (t
            (cons(remove nil(mapcar(lambda(elemento)(if(funcall (car f_cond) elemento)elemento nil))l)))
                    (apply #'clasificar_valores l (cdr f_cond))
        )
    )
)

;Tomando la idea del ejercicio primeroCentralyUltimo. Sin definir otras funciones adicionales, implementar una función en lisp llamada extraer_extremos que, dado un número arbitrario de listas suministradas como parámetros, devuelva una lista de listas donde cada sublista contenga exactamente dos elementos: el primer y el último elemento de la lista original. Si la lista tiene un solo elemento, este se repetirá. Si está vacía, devolverá (NIL NIL).
;(extraer_extremos '(1 2 3) '(9) '() '(5 6 7 8)) → ((1 3) (9 9) (NIL NIL) (5 8))

;Version mapcar
(defun extraer_extremos (&rest l)
    (mapcar(lambda(sublista)(list (car sublista)(car(reverse sublista))))l)
)

;Version recursiva
(defun extraer_extremos (&rest l)
    (cond
        ((null l)(list nil nil))
        (t
            (cons(list
                (car(car l))(car(reverse(car l)))
                    (apply #'extraer_extremos (cdr l))
            ))
        )
    )
)

;Sin definir otras funciones adicionales, implementar una función en lisp llamada procesar_mixto que reciba una función unaria y un número arbitrario de argumentos que pueden ser números o listas. La función debe devolver una nueva lista plana donde se haya aplicado la función suministrada directamente sobre el argumento (si este es un número), o sobre la longitud de la lista (si el argumento es una lista).  
;(procesar_mixto #'1+ 5 '(A B C) 10 NIL '(X Y)) -> (6 4 11 1 3)

;Version mapcar
(defun procesar_mixto (f &rest a)
    (mapcar(lambda(x)(if(listp x)(funcall f (length x))(funcall f x)))a)
)

;Version recursiva
(defun procesar_mixto (f &rest a)
    (cond
        ((null a) nil)
        ((listp (car a))
            (cons(funcall f (length (car a)))(apply #'procesar_mixto f (cdr a)))
        )
        (t
            (cons(funcall f (car a))(apply #'procesar_mixto f (cdr a)))
        )
    )
)   

;Sin definir otras funciones adicionales y sin usar llamadas recursivas, implementar una función en LISP llamada operar_extremos que, dada una función binaria (como #'+ o #'*) y un número arbitrario de listas no vacías, devuelva una lista plana. Cada elemento de esta lista devuelta será el resultado de aplicar la función binaria entre el primer y el último elemento de la sublista correspondiente.
;(operar_extremos #'+ '(1 2 3) '(10 20) '(5 5 5 5)) -> (4 30 10) 
(defun operar_extremos (f_bin &rest l)
    (mapcar(lambda(sublista)(funcall f_bin (car sublista) (car(reverse sublista))))l)
)

;Sin definir otras funciones adicionales y sin utilizar la función mapcar, implementar una función recursiva en LISP llamada filtrar_tipos que, dado un predicado (una función que devuelve T o NIL, como #'numberp o #'listp) y un número arbitrario de argumentos mezclados, devuelva una única lista plana que contenga exclusivamente los argumentos que cumplen dicho predicado.
;(filtrar_tipos #'numberp 1 '(A) 2 NIL 3) -> (1 2 3)
(defun filtrar_tipos (f_pred &rest a)
    (cond
        ((null a) nil)
        ((funcall f_pred (car a))
            (cons(car a)(apply #'filtrar_tipos f_pred (cdr a)))
        )
        (t
            (apply #'filtrar_tipos f_pred (cdr a))
        )
    )
)

;Sin definir otras funciones adicionales y utilizando únicamente mapcar (sin recursividad), implementar una función en LISP llamada cruzar_funciones que, dado un valor cualquiera y un número arbitrario de funciones, devuelva una lista de pares. Cada "par" será una sublista de dos elementos: el valor original y el resultado de aplicarle la función correspondiente.
;(cruzar_funciones 5 #'1+ #'1- (lambda (x) (* x 2))) -> ((5 6) (5 4) (5 10))
(defun cruzar_funciones (n &rest f)
    (mapcar(lambda(funcion)(list n (funcall funcion n)))f)
)

;Sin definir otras funciones adicionales y sin utilizar la función mapcar, implementar una función puramente recursiva en LISP llamada suma_estricta que, dado un número arbitrario de argumentos de distintos tipos (números, listas, átomos nulos), devuelva la suma matemática exclusivamente de los argumentos que sean números, ignorando por completo el resto. En caso de no haber ningún número, debe devolver 0.
;(suma_estricta 10 '(A B) 5 NIL 20) -> 35
;(suma_estricta '(1 2) NIL) -> 0
(defun suma_estricta (&rest a)
    (cond
        ((null a) 0)
        ((numberp (car a))
            (+ (car a)(apply #'suma_estricta (cdr a)))
        )
        (t
            (apply #'suma_estricta (cdr a))
        )
    )
)

;Sin definir otras funciones y usando exclusivamente recursividad pura (prohibido mapcar, flatten o similares), implementa la función sumar_profundidad.
;Esta función recibe una función de predicado (ej. #'evenp, #'oddp) y una lista que puede contener números, símbolos, o cualquier cantidad de listas anidadas a cualquier nivel. Debe devolver la suma matemática de todos los números en cualquier nivel de profundidad que cumplan el predicado. Si no hay ninguno, devuelve 0.
;(sumar_profundidad #'evenp '(1 (2 A) ((3 4) NIL 5) B (6))) -> 12  ;; (Suma de 2 + 4 + 6)
(defun sumar_profundidad (f_pred l)
    (cond
        ((null l) 0)
        ((listp (car l))
            (apply #'sumar_profundidad f_pred (car l))(apply #'sumar_profundidad f_pred (cdr l))
        )
        (t
            (if((funcall f_pred (car l))
                (+(car l)(apply #'sumar_profundidad f_pred (cdr l))))
                (apply #'sumar_profundidad f_pred (cdr l)))
        )
    )
)

;Sin usar recursividad y utilizando únicamente mapcar, implementa la función batalla_funciones.
;Recibe tres argumentos: una lista de valores numéricos, una lista de predicados, y una lista de funciones transformadoras. (Asume que las tres listas tienen la misma longitud).
;La función debe evaluar cada valor con su predicado correspondiente en la misma posición. Si el predicado devuelve T, se le aplica la transformación correspondiente. Si devuelve NIL, el valor se queda igual. Devuelve la lista resultante.
;(batalla_funciones '(2 3 4) '(#'evenp #'evenp #'oddp) '(#'1+ #'1- (lambda (x) (* x 10)))) -> (3 3 4)
(defun batalla_funciones (l_val l_pred l_func)
    (mapcar (lambda (v p f)(if (funcall p v))(funcall f v) v) l_val l_pred l_func)
)

;Sin usar llamadas recursivas y usando puramente lógica funcional (mapcar, remove, apply...), implementa la función limpieza_estructural.
;Recibe un predicado y un número arbitrario de listas mixtas (pueden contener números y otras cosas). Debe devolver una lista de listas donde, de cada sublista, se hayan eliminado los elementos que NO cumplen el predicado.
;La trampa mortal: Si tras el filtrado una sublista queda completamente vacía, NO debe aparecer en el resultado final.
;(limpieza_estructural #'numberp '(A 1 B 2) '(X Y Z) '(3 NIL 4)) -> ((1 2) (3 4)) 
(defun limpieza_estructural (f &rest l)
    (remove nil(mapcar (lambda(sublista)(remove nil(mapcar (lambda(x)(if(funcall f x)x nil))sublista))l)))
)

;Basado en tus ejercicios aplanar y filtrar, vamos a dar un paso más allá.
;Implementa una función llamada (recolectar_si predicado arbol) que reciba una función unaria que devuelva un booleano (predicado) y una lista que puede tener cualquier nivel de anidamiento (arbol). La función debe devolver una lista plana (sin anidar) con todos los átomos del árbol para los cuales la función predicado devuelva T.
;(recolectar_si #'evenp '(1 (2 3 (4)) 5 (6 (7 8)))) ->(2 4 6 8)
(defun recolectar_si (f l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (append (recolectar_si f (car l))(recolectar_si f (cdr l)))
        )
        ((funcall f (car l))
            (cons (car l)(recolectar_si f (cdr l)))
        )
        (t
            (recolectar_si f (cdr l))
        )
    )
)

;Implementa una función llamada (pipeline x &rest funciones) que reciba un valor inicial x y un número arbitrario de funciones unarias. La función debe aplicar las funciones en cadena de izquierda a derecha. Es decir, el resultado de la primera función se le pasa a la segunda, y así sucesivamente: fn(...f2(f1(x))...)fn(...f 2(f 1​	(x))...). Si no se le pasa ninguna función, debe devolver x
;(pipeline 3 (lambda (x) (+ x 2)) (lambda (x) (* x 2)) #'1-) -> 9   ;; Explicación: 3+2 = 5 -> 5*2 = 10 -> 10-1 = 9
(defun pepeline (x &rest funciones)
    (cond
        ((null funciones) x)
        (t
            (apply #'pepeline (funcall (car funciones) x) (cdr funciones))
        )
    )
)

;Implementa ahora la función (diagonal M) que, dada una matriz cuadrada M, devuelva una lista plana con los elementos de su diagonal principal.
;(Pista: Seguramente necesites una función auxiliar que lleve cuenta de la posición, o jugar con la recursividad extrayendo el car de la primera fila y avanzando en las siguientes).
(defun diagonal (m)
    (cond
        ((null m) nil)
        (t
            (cons (caar m)(diagonal(mapcar #'cdr (cdr m))))  
        )
    )
)

;Implementa una función (producto_cartesiano A B) que reciba dos listas planas (que representan conjuntos) y devuelva una lista de listas con todas las parejas posibles formadas por un elemento de A y un elemento de B.
(defun producto_cartesiano (a b)
   (apply #append(mapcar (lambda (x)(mapcar (lambda (y)((list x y))b)))a))
)

;Implementa una función (duplicar_elementos lista) que reciba una lista plana y devuelva otra lista donde cada elemento aparezca dos veces seguidas.
(defun duplicar_elementos (l)
    (apply #append (mapcar (lambda(x)(list x x))l))
)

;Implementa tu propia versión de la función nativa reverse, llamándola (mi_reverso lista). Debe coger una lista y devolverla dada la vuelta usando recursividad pura.
(defun mi_reverso (l)
    (cond
        ((null l) nil)
        (t
            (append (mi_reverso(cdr l)(list (car l))))
        )
    )
)

;Implementa una función (sustituir_si predicado nuevo_valor arbol) que reciba una función unaria (predicado), un valor y una lista con cualquier nivel de anidamiento (arbol). La función debe devolver el mismo árbol intacto en su estructura, pero sustituyendo por el nuevo_valor todos los átomos para los cuales la función predicado devuelva T.
(defun sustituir_si (f v l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (sustituir_si f v (car l))(sustituir_si f v (cdr l)))
        )
        ((funcall f (car l))
            (cons v (sustituir_si f v (cdr l)))
        )
        (t
            (cons (car l)(sustituir_si f v (cdr l)))
        )
    )
)

;Implementa una función (multi_mapcar lista &rest funciones) que reciba una lista plana y un número arbitrario de funciones unarias. La función debe devolver una lista de listas, donde cada sublista contenga el resultado de aplicar todas y cada una de las funciones al elemento correspondiente de la lista original.
(defun multi_mapcar (l &rest f)
    (mapcar (lambda(x)(mapcar (lambda(y)(funcall y x))f))l)
)

;Implementa una función (sumar_columnas matriz) que reciba una matriz (lista de listas) y devuelva una lista plana con la suma matemática de los elementos de cada columna. Prohibido usar funciones auxiliares; resuélvelo mediante recursividad o explotando mapcar y apply.
(defun sumar_columnas (m)
    (apply #'mapcar #'+ m)
)

;Implementa la función (comprimir lista) que reciba una lista plana y devuelva una lista de listas, donde cada sublista contenga el elemento y el número de veces consecutivas que aparece en la lista original.
(defun comprimir (l)
    (cond
        ((null l) nil)
        ((= (car l) (caar(comprimir(cdr l))))
            (cons(list (car l (1+ (cadar)(comprimir (cdr l)))))
                    (comprimir (cdr l)))
        )
        (t
            (cons (list (car l) 1)(comprmir (cdr l)))
        )
    )
)

;Implementa una función (map_arbol f arbol) que reciba una función unaria f y una lista con cualquier nivel de anidamiento arbol. La función debe devolver un árbol con la misma estructura exacta, pero con la función aplicada a cada uno de sus átomos.
(defun map_arbol (f l)
    (cond
        ((null l) nil)
        ((listp (car l))
            (cons (map_arbol f (car l))(map_arbol f (cdr l)))
        )
        (t
            (cons (funcall f (car l))(map_arbol f (cdr l)))
        )
    )
)

;Basado en el Ejercicio 2 de tu Práctica 1. Implementa una función (aplicar_alterno f g &rest valores) que reciba dos funciones unarias y una cantidad variable de valores. Debe devolver una lista plana aplicando f al primer valor, g al segundo, f al tercero, g al cuarto, y así sucesivamente.
(defun aplicar_alterno (f g &rest valores)
    (cond
        ((null valores) nil)
        (t
            (cons (funcall f (car valores))(apply #'aplicar_alterno g f (cdr valores)))
        )
    )
)

;En tu Práctica 1 te piden implementar la matriz transpuesta. Una matriz transpuesta convierte las filas en columnas y viceversa. Implementa (transpuesta m) sin usar recursividad explícita
(defun traspuesta (m)
    (apply #'mapcar #'list m)
)

;Implementa la función (interseccion A B) que reciba dos listas planas que representan conjuntos, y devuelva una lista plana solo con los elementos de A que también estén presentes en B. Prohibido usar funciones auxiliares o la función nativa de intersección.
(defun interseccion (a b)
    (cond
        ((OR (null a)(null b)) nil)
        ((member (car a) b)
            (list (car a)(interseccion (cdr a) b))
        )
        (t
            (interseccion (cdr a) b)
        )
    )
)
