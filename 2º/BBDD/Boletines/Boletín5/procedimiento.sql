--creamos un procedimiento que nos suelte el script xml con datos de la bbdd
create or replace procedure ventas_xml is 
    --declaramos los cursores que recorreran los pedidos y la lineas de cada pedido
    cursor pedidos is 
        select p.idpedido, p.fechapedido, p.fechaentrega, p.direccionentrega, p.empleadoventa, c.idcliente, c.nombre, c.apellidos, c.email, c.direccionpostal, c.ciudad
        from pedidos p join clientes c on p.idcliente = c.idcliente;

    cursor lineas (p_idpedido number) is
        select prod.idproducto, prod.idproveedor, prod.descripcion, prod.precio_compra, prod.precio_venta, prod.stock, pp.cantidad, pp.precioreal
        from pedidos_productos pp join productos prod on pp.idproducto = prod.idproducto
        where pp.idpedido = p_idpedido;
begin
    --cabecera, enlace externo a dtd y abrimos raiz
    dbms_output.put_line('<?xml version="1.0" encoding="UTF-8"?>');
    dbms_output.put_line('<!DOCTYPE lista_ventas SYSTEM "ventas.dtd">');
    dbms_output.put_line('<lista_ventas>');
    for reg_pedido in pedidos loop  
        --conectamos los datos del pedido con los atributos obtenidos
        dbms_output.put_line('    <pedido idpedido="' || reg_pedido.idpedido || '" fechapedido="' || to_char(reg_pedido.fechapedido, 'dd/mm/yyyy') || '" fechaentrega="' || to_char(reg_pedido.fechaentrega, 'dd/mm/yyyy') || '" direccionentrega="' || reg_pedido.direccionentrega || '" empleadoventa="' || reg_pedido.empleadoventa || '">');
        --conectamos los datos del clinte con los atributos de la misma manera que antes
        dbms_output.put_line('        <cliente idcliente="' || reg_pedido.idcliente || '">');
        dbms_output.put_line('            <nombre>' || reg_pedido.nombre || '</nombre>');
        dbms_output.put_line('            <apellidos>' || reg_pedido.apellidos || '</apellidos>');
        dbms_output.put_line('            <email>' || reg_pedido.email || '</email>');
        dbms_output.put_line('            <direccionpostal>' || reg_pedido.direccionpostal || '</direccionpostal>');
        dbms_output.put_line('            <ciudad>' || reg_pedido.ciudad || '</ciudad>');
        dbms_output.put_line('        </cliente>');
        --abrimos las lineas de pedidos
        dbms_output.put_line('        <lineas_pedido>');
        for reg_linea in lineas(reg_pedido.idpedido) loop
        --lo mismo que antes pero ahora, para cada linea de pedido la informacion de los productos y de la venta
        dbms_output.put_line('            <producto idproducto="' || reg_linea.idproducto || '" idproveedor="' || reg_linea.idproveedor || '">');
        dbms_output.put_line('                <descripcion>' || reg_linea.descripcion || '</descripcion>');
        dbms_output.put_line('                <precio_compra>' || reg_linea.precio_compra || '</precio_compra>');
        dbms_output.put_line('                <precio_venta>' || reg_linea.precio_venta || '</precio_venta>');
        dbms_output.put_line('                <stock>' || reg_linea.stock || '</stock>');
        dbms_output.put_line('                <cantidad>' || reg_linea.cantidad || '</cantidad>');
        dbms_output.put_line('                <precioreal>' || reg_linea.precioreal || '</precioreal>');
        dbms_output.put_line('            </producto>');
        end loop;
        --cerramos las lineas y el pedido
        dbms_output.put_line('        </lineas_pedido>');
        dbms_output.put_line('    </pedido>');
    end loop;
    --cerramos raiz
    dbms_output.put_line('</lista_ventas>');
end;