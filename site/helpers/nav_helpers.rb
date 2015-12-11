module NavHelpers
  def docs_in_order
    docs_list = []
    data.docs_index.docs.each do |doc|
      docs_list << { title: doc.title, path: doc.path }
      next if doc.sub_docs.nil?

      doc.sub_docs.each do |sub_doc|
        docs_list << { category: doc.title, title: sub_doc.title, path: sub_doc.path }
        next if sub_doc.sub_docs.nil?

        sub_doc.sub_docs.each do |sub_sub_doc|
          docs_list << { category: doc.title, sub_doc: sub_doc.title, title: sub_sub_doc.title, path: sub_sub_doc.path }
        end
      end
    end
    docs_list
  end

  def get_prev_doc(current_article_path)
    docs = docs_in_order
    index = docs_in_order.find_index { |d| d[:path] == current_article_path }

    if index == 0
      nil
    else
      docs[index - 1]
    end
  end

  def get_next_doc(current_article_path)
    docs = docs_in_order
    index = docs_in_order.find_index { |d| d[:path] == current_article_path }

    if index == docs.count - 1
      nil
    else
      docs[index + 1]
    end
  end
end
